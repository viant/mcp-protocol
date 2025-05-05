package schema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// InputSchema is a JSON InputSchema representation of a struct.

// schemaForTypeInternal returns a JSON InputSchema representation for a given reflect.Type.
// The inSlice flag is used to determine if we are processing an element inside a slice.
func schemaForTypeInternal(t reflect.Type, inSlice bool) map[string]interface{} {
	schema := make(map[string]interface{})

	// Special handling for time.Time: treat as ISO 8601 string.
	if t == reflect.TypeOf(time.Time{}) {
		schema["type"] = "string"
		schema["format"] = "date-time"
		return schema
	}

	// Handle pointer types.
	if t.Kind() == reflect.Ptr {
		// Unwrap pointer.
		schema = schemaForTypeInternal(t.Elem(), inSlice)
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
	case reflect.Slice, reflect.Array:
		schema["type"] = "array"
		// When processing slice items, set inSlice = true.
		schema["items"] = schemaForTypeInternal(t.Elem(), true)
	case reflect.Map:
		schema["type"] = "object"
		// For maps, set additionalProperties based on the element type.
		schema["additionalProperties"] = schemaForTypeInternal(t.Elem(), false)
	case reflect.Struct:
		// For structs, recursively convert their fields.
		schema["type"] = "object"
		properties, required := structToProperties(t)
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

// schemaForType is a wrapper that starts schema generation with inSlice=false.
func schemaForType(t reflect.Type) map[string]interface{} {
	return schemaForTypeInternal(t, false)
}

// structToProperties converts a struct type into MCP InputSchema properties and required fields.
func structToProperties(t reflect.Type) (ToolInputSchemaProperties, []string) {
	properties := make(ToolInputSchemaProperties)
	var required []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Only process exported fields.
		if !field.IsExported() {
			continue
		}

		// Parse struct tags for json and format.
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
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
		fieldSchema := schemaForType(field.Type)
		// Apply format tag if present (e.g., date format).
		if formatTag := field.Tag.Get("format"); formatTag != "" {
			fieldSchema["format"] = formatTag
		}
		// Set the property in overall schema.
		properties[fieldName] = fieldSchema
		if field.Type.Kind() != reflect.Ptr && !omitempty {
			required = append(required, fieldName)
		}
	}

	return properties, required
}

func (s *ToolInputSchema) Load(v any) error {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("expected a struct type, got %s", t.Kind())
	}
	properties, required := structToProperties(t)
	s.Properties = properties
	s.Required = required
	s.Type = "object"
	return nil
}

// CallToolRequestParams is a structure that holds the parameters for a tool call request.
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
