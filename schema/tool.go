package schema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"
)

// visitedTypes tracks types currently being processed to detect cyclic definitions.
var (
	visitedMu    sync.Mutex
	visitedTypes map[reflect.Type]bool
)

// buildJSONSchema constructs a JSON-schema fragment that represents the supplied Go reflect.Type.
// The `inSlice` flag is an internal recursion marker: it is true when the type
// currently being processed is the *element* of a slice/array. That allows the
// algorithm to decide whether automatically added `nullable:true` (for pointer
// types) should be kept or suppressed.
// It detects and breaks cycles in types, returning an empty object schema on recursion.
func buildJSONSchema(t reflect.Type, inSlice bool, opts ...StructToPropertiesOption) map[string]interface{} {
	schema := make(map[string]interface{})

	// initialize cycle-detection state on outermost call
	visitedMu.Lock()
	first := visitedTypes == nil
	if first {
		visitedTypes = make(map[reflect.Type]bool)
	}
	// detect cyclic types: if this type is already being processed, break the cycle
	if visitedTypes[t] {
		visitedMu.Unlock()
		if first {
			visitedMu.Lock()
			visitedTypes = nil
			visitedMu.Unlock()
		}
		return map[string]interface{}{"type": "object"}
	}
	visitedTypes[t] = true
	visitedMu.Unlock()

	defer func() {
		visitedMu.Lock()
		delete(visitedTypes, t)
		if first {
			visitedTypes = nil
		}
		visitedMu.Unlock()
	}()

	// Special handling for time.Time: treat as ISO 8601 string.
	if t == reflect.TypeOf(time.Time{}) {
		schema["type"] = "string"
		schema["format"] = "date-time"
		return schema
	}

	// Handle pointer types.
	if t.Kind() == reflect.Ptr {
		// Unwrap pointer.
		schema = buildJSONSchema(t.Elem(), inSlice, opts...)
		// Mark as nullable unless we are processing a slice element.
		if !inSlice {
			schema["nullable"] = true
		}
		return schema
	}

	switch t.Kind() {
	case reflect.Bool:
		schema["type"] = "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema["type"] = "integer"
	case reflect.Float32, reflect.Float64:
		schema["type"] = "number"
	case reflect.String:
		schema["type"] = "string"
	case reflect.Interface:
		// Unconstrained value – same representation the MCP spec uses.
		return schema
	case reflect.Slice, reflect.Array:
		schema["type"] = "array"
		// When processing slice items, set inSlice = true.
		schema["items"] = buildJSONSchema(t.Elem(), true, opts...)
	case reflect.Map:
		schema["type"] = "object"
		// Open-ended objects when the map value is interface{}.
		if t.Elem().Kind() == reflect.Interface {
			schema["additionalProperties"] = true
		} else {
			schema["additionalProperties"] = buildJSONSchema(t.Elem(), false, opts...)
		}
	case reflect.Struct:
		// For structs, recursively convert their fields.
		schema["type"] = "object"
		properties, required := StructToProperties(t, opts...)
		schema["properties"] = properties
		if len(required) > 0 {
			schema["required"] = required
		}
	default:
		// Fallback to string type.
		schema["type"] = "string"
	}
	return schema
}

// typeSchema is a package-internal helper that starts schema generation with
// `inSlice == false` (i.e. at the “top level”). Callers inside this package
// should use this instead of calling buildJSONSchema directly so they do not need to
// know about the recursion flag.
// use. It hides the recursion flag (`inSlice`) by always starting the walk in
// "top-level" mode (inSlice == false).
func typeSchema(t reflect.Type, opts ...StructToPropertiesOption) map[string]interface{} {
	// typeSchema hides the recursion flag for callers at the top level.
	return buildJSONSchema(t, false, opts...)
}

// StructToPropertiesOption defines an option for controlling field inclusion/exclusion in StructToProperties.
type StructToPropertiesOption func(*structToPropertiesOptions)

type structToPropertiesOptions struct {
	SkipFieldHook  func(field reflect.StructField) bool
	IsRequiredHook func(field reflect.StructField) bool
	FormatHook     func(field reflect.StructField) string
	NullableHook   func(field reflect.StructField) *bool
}

// WithSkipFieldHook returns an option to set a hook for skipping fields based on reflect.StructField.
func WithSkipFieldHook(hook func(field reflect.StructField) bool) StructToPropertiesOption {
	return func(o *structToPropertiesOptions) {
		o.SkipFieldHook = hook
	}
}

// WithIsRequiredHook returns an option to set a hook for determining if a field is required based on reflect.StructField.
func WithIsRequiredHook(hook func(field reflect.StructField) bool) StructToPropertiesOption {
	return func(o *structToPropertiesOptions) {
		o.IsRequiredHook = hook
	}
}

// WithFormatHook returns an option to set a hook for overriding field format based on reflect.StructField.
func WithFormatHook(hook func(field reflect.StructField) string) StructToPropertiesOption {
	return func(o *structToPropertiesOptions) {
		o.FormatHook = hook
	}
}

// WithNullableHook returns an option to explicitly set or unset `nullable` for
// a given struct field. The hook must return a pointer to bool – *true* to
// force `nullable:true`, *false* to force it off, and *nil* to keep the default
// behaviour (automatic for pointer types outside slices).
func WithNullableHook(hook func(field reflect.StructField) *bool) StructToPropertiesOption {
	return func(o *structToPropertiesOptions) {
		o.NullableHook = hook
	}
}

// StructToProperties converts a struct type into MCP InputSchema properties and required fields.
// It accepts optional StructToPropertiesOption to customize behavior (e.g., skipping fields, required logic, format overrides).
func StructToProperties(t reflect.Type, opts ...StructToPropertiesOption) (ToolInputSchemaProperties, []string) {
	var opt structToPropertiesOptions
	for _, o := range opts {
		o(&opt)
	}

	properties := make(ToolInputSchemaProperties)
	var required []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if opt.SkipFieldHook != nil && opt.SkipFieldHook(field) {
			continue
		}
		// Only process exported fields.
		if !field.IsExported() {
			continue
		}
		// Parse struct tags for json and format.
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		isInternal := field.Tag.Get("internal")
		if isInternal != "" {
			continue
		}
		var fieldName string
		var omitempty bool
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" {
				fieldName = parts[0]
			}
			for _, opt := range parts[1:] {
				if opt == "omitempty" {
					omitempty = true
				}
			}
		}

		if fieldName == "" {
			fieldName = field.Name
		}
		// Generate the field's JSON schema.
		fieldSchema := typeSchema(field.Type, opts...)
		// Determine format via hook or tag.
		var fmtVal string
		if opt.FormatHook != nil {
			fmtVal = opt.FormatHook(field)
		} else {
			fmtVal = field.Tag.Get("format")
		}
		if fmtVal != "" {
			fieldSchema["format"] = fmtVal
		}

		if desc := field.Tag.Get("description"); desc != "" {
			fieldSchema["description"] = desc
		}

		if choice := field.Tag.Get("choice"); choice != "" {
			re := regexp.MustCompile(`choice:"([^"]+)"`)
			matches := re.FindAllStringSubmatch(string(field.Tag), -1)
			var choices []string
			for _, match := range matches {
				choices = append(choices, match[1])
			}
			fieldSchema["enum"] = choices
		}

		// Apply nullable override via hook if provided.
		if opt.NullableHook != nil {
			if nullablePtr := opt.NullableHook(field); nullablePtr != nil {
				if *nullablePtr {
					fieldSchema["nullable"] = true
				} else {
					delete(fieldSchema, "nullable")
				}
			}
		}
		// Set the property in overall schema.
		properties[fieldName] = fieldSchema
		tag := string(field.Tag)
		isOptional := strings.Contains(tag, "required=false") || strings.Contains(tag, "optional")
		if isOptional {
			continue
		}
		isRequiredValue, isRequired := field.Tag.Lookup("required")
		isRequired = isRequiredValue == "true" || (isRequiredValue == "" && isRequired)
		if (field.Type.Kind() != reflect.Ptr && !omitempty) || isRequired {
			required = append(required, fieldName)
		}
	}

	return properties, required
}

func (s *ToolInputSchema) Load(v any, options ...StructToPropertiesOption) error {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("expected a struct type, got %s", t.Kind())
	}
	properties, required := StructToProperties(t, options...)
	s.Properties = properties
	s.Required = required
	s.Type = "object"
	return nil
}

func (s *ToolOutputSchema) Load(v any, options ...StructToPropertiesOption) error {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("expected a struct type, got %s", t.Kind())
	}
	properties, required := StructToProperties(t, options...)
	s.Properties = properties
	s.Required = required
	s.Type = "object"
	return nil
}

// NewCallToolRequestParams creates a new CallToolRequestParams instance from a command struct.
func NewCallToolRequestParams[T any](name string, cmd T) (*CallToolRequestParams, error) {
	results := &CallToolRequestParams{Name: name, Arguments: map[string]interface{}{}}
	data, err := json.Marshal(cmd)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &results.Arguments)
	if err != nil {
		return nil, err
	}
	return results, nil
}
