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
	Annotate func(string, reflect.StructField) (string, any)
}
type Request struct {
	Type     reflect.Type
	Path     string
	Property string
}

func (r Reflector) Compile(request Request) (map[string]any, error) {
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
	if t == reflect.TypeFor[time.Time]() {
		return map[string]any{"type": "string", "format": "date-time"}, nil
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
			if field.Field.Tag.Get("setMarker") == "true" {
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
