// Package jsonschema compiles native MCP JSON schemas from canonical type and
// resource owners. It never inspects invocation values or executes user code.
package jsonschema

import (
	"crypto/sha256"
	"encoding"
	"encoding/json"
	"fmt"
	xshape "github.com/viant/x/shape"
	"reflect"
	"strings"
	"time"
)

// Reflector uses X's encoding/json field projection; annotations cannot change
// selection, Go identity, JSON names, or source types.
type Reflector struct {
	// WireTypes declares owner-authored canonical wire representations for custom
	// JSON types. It never infers a contract or invokes custom marshal methods.
	WireTypes map[reflect.Type]reflect.Type

	// ExcludeInternal removes internal-tagged JSON winners from public input schemas.
	// The zero value retains ordinary encoding/json schema semantics.
	ExcludeInternal bool
	Annotate        func(string, reflect.StructField) (string, any)
}
type Request struct {
	Type     reflect.Type
	Path     string
	Property string
}

func (r Reflector) Compile(request Request) (map[string]any, error) {
	if err := validateWireTypes(r.WireTypes); err != nil {
		return nil, err
	}
	prefix := "#/$defs/"
	if request.Property != "" {
		prefix = "#/properties/" + strings.ReplaceAll(strings.ReplaceAll(request.Property, "~", "~0"), "/", "~1") + "/$defs/"
	}
	compiler := reflectionCompiler{reflector: r, prefix: prefix, active: map[reflect.Type]string{}, objects: map[string]map[string]any{}, referenced: map[string]bool{}}
	result, err := compiler.value(request.Type, request.Path)
	if err != nil {
		return nil, err
	}
	if len(compiler.referenced) > 0 {
		root := map[string]any{}
		for key, value := range result {
			root[key] = value
		}
		definitions := map[string]any{}
		for key := range compiler.referenced {
			definitions[key] = compiler.objects[key]
		}
		root["$defs"] = definitions
		result = root
	}
	return result, nil
}

type reflectionCompiler struct {
	reflector  Reflector
	prefix     string
	active     map[reflect.Type]string
	objects    map[string]map[string]any
	referenced map[string]bool
}

func (c *reflectionCompiler) value(t reflect.Type, path string) (map[string]any, error) {
	if t == nil {
		return nil, fmt.Errorf("JSON schema requires source type")
	}
	if t.Kind() == reflect.Pointer {
		item, err := c.value(t.Elem(), path)
		if err != nil {
			return nil, err
		}
		if kind, ok := item["type"].(string); ok {
			copy := map[string]any{}
			for k, v := range item {
				copy[k] = v
			}
			copy["type"] = []string{kind, "null"}
			return copy, nil
		}
		return map[string]any{"anyOf": []any{item, map[string]any{"type": "null"}}}, nil
	}
	if wire := c.reflector.WireTypes[t]; wire != nil {
		return c.value(wire, path)
	}
	if t == reflect.TypeFor[time.Time]() {
		return map[string]any{"type": "string", "format": "date-time"}, nil
	}
	if t == reflect.TypeFor[json.RawMessage]() {
		// RawMessage represents any JSON value, regardless of its underlying slice type.
		return map[string]any{}, nil
	}
	pointer := reflect.PointerTo(t)
	if t.Implements(reflect.TypeFor[json.Marshaler]()) || pointer.Implements(reflect.TypeFor[json.Marshaler]()) ||
		pointer.Implements(reflect.TypeFor[json.Unmarshaler]()) || t.Implements(reflect.TypeFor[encoding.TextMarshaler]()) ||
		pointer.Implements(reflect.TypeFor[encoding.TextMarshaler]()) || pointer.Implements(reflect.TypeFor[encoding.TextUnmarshaler]()) {
		return nil, fmt.Errorf("custom JSON source %s requires authored wire authority", t)
	}
	switch t.Kind() {
	case reflect.Struct:
		if id := c.active[t]; id != "" {
			c.referenced[id] = true
			return map[string]any{"$ref": c.prefix + id}, nil
		}
		expression, err := (xshape.Resolver{}).Expression(t)
		if err != nil {
			return nil, err
		}
		id := fmt.Sprintf("T_%x", sha256.Sum256([]byte(t.PkgPath()+":"+expression+":"+path)))
		c.active[t] = id
		defer delete(c.active, t)
		properties := map[string]any{}
		result := map[string]any{"type": "object", "properties": properties}
		c.objects[id] = result
		fields, err := xshape.Linked(t).JSONFields()
		if err != nil {
			return nil, err
		}
		for _, field := range fields {
			if field.Field.Tag.Get("setMarker") == "true" || (c.reflector.ExcludeInternal && internalJSONField(t, field.Field.Index)) {
				continue
			}
			fieldPath := field.Field.Name
			if path != "" {
				fieldPath = path + "." + fieldPath
			}
			property, err := c.value(field.Field.ReflectedType, fieldPath)
			if err != nil {
				return nil, fmt.Errorf("field %s: %w", fieldPath, err)
			}
			if field.Quoted {
				property = map[string]any{"type": "string"}
			}
			if c.reflector.Annotate != nil {
				description, example := c.reflector.Annotate(fieldPath, field.Field.StructField())
				if description != "" {
					property["description"] = description
				}
				if example != nil {
					property["examples"] = []any{example}
				}
			}
			properties[field.Name] = property
		}
		return result, nil
	case reflect.Slice, reflect.Array:
		if t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8 {
			return map[string]any{"type": "string", "contentEncoding": "base64"}, nil
		}
		element, err := c.value(t.Elem(), path)
		if err != nil {
			return nil, err
		}
		return map[string]any{"type": "array", "items": element}, nil
	case reflect.Map:
		if t.Key().Kind() != reflect.String {
			return nil, fmt.Errorf("map key %s is unsupported", t.Key())
		}
		value, err := c.value(t.Elem(), path)
		if err != nil {
			return nil, err
		}
		return map[string]any{"type": "object", "additionalProperties": value}, nil
	case reflect.Bool:
		return map[string]any{"type": "boolean"}, nil
	case reflect.String:
		return map[string]any{"type": "string"}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]any{"type": "integer"}, nil
	case reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}, nil
	case reflect.Interface:
		if t.NumMethod() == 0 {
			return map[string]any{}, nil
		}
	}
	return nil, fmt.Errorf("source type %s is unsupported", t)
}

// internalJSONField checks the canonical selected index, including promoted
// embedding owners. Filtering happens after JSON dominance so a shadowed public
// field never becomes visible when an internal winner is excluded.
func internalJSONField(owner reflect.Type, index []int) bool {
	for _, position := range index {
		for owner.Kind() == reflect.Pointer {
			owner = owner.Elem()
		}
		field := owner.Field(position)
		if field.Tag.Get("internal") == "true" {
			return true
		}
		owner = field.Type
	}
	return false
}

func validateWireTypes(authority map[reflect.Type]reflect.Type) error {
	for source, wire := range authority {
		if source == nil || wire == nil || source.Kind() == reflect.Pointer || wire.Kind() == reflect.Pointer {
			return fmt.Errorf("wire authority requires non-pointer source and representation types")
		}
		seen := map[reflect.Type]bool{}
		for current := source; current != nil; current = authority[current] {
			if seen[current] {
				return fmt.Errorf("cyclic authored wire authority for %s", source)
			}
			seen[current] = true
		}
		pointer := reflect.PointerTo(source)
		if !source.Implements(reflect.TypeFor[json.Marshaler]()) && !pointer.Implements(reflect.TypeFor[json.Marshaler]()) && !pointer.Implements(reflect.TypeFor[json.Unmarshaler]()) && !source.Implements(reflect.TypeFor[encoding.TextMarshaler]()) && !pointer.Implements(reflect.TypeFor[encoding.TextMarshaler]()) && !pointer.Implements(reflect.TypeFor[encoding.TextUnmarshaler]()) {
			return fmt.Errorf("wire authority requires a custom JSON source: %s", source)
		}
	}
	return nil
}
