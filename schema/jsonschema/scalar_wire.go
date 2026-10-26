package jsonschema

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/viant/mcp-protocol/schema/fieldname"
)

// ScalarWireSchema reads an owner-authored wireSchema tag for a custom JSON
// scalar. It never invokes serialization methods or changes runtime binding.
// Callers must apply their ordinary field visibility policy first.
// Supported declarations contain type (a scalar name or union) and optional
// format. Object graphs, references and arbitrary schema overrides are excluded.
func ScalarWireSchema(field reflect.StructField) (map[string]any, bool, error) {
	raw, found := field.Tag.Lookup("wireSchema")
	if !found {
		return nil, false, nil
	}
	fail := func(message string) (map[string]any, bool, error) {
		return nil, true, fmt.Errorf("field %s wireSchema: %s", field.Name, message)
	}
	base := field.Type
	for base.Kind() == reflect.Pointer {
		base = base.Elem()
	}
	custom := fieldname.CustomJSON(base)
	for _, contract := range []reflect.Type{reflect.TypeFor[json.Marshaler](), reflect.TypeFor[encoding.TextMarshaler]()} {
		custom = custom || base.Implements(contract) || reflect.PointerTo(base).Implements(contract)
	}
	if base == reflect.TypeFor[time.Time]() || !custom {
		return fail("requires a custom JSON scalar")
	}
	for _, option := range strings.Split(field.Tag.Get("json"), ",")[1:] {
		if option == "string" {
			return fail("conflicts with quoted JSON encoding")
		}
	}
	decoder := json.NewDecoder(bytes.NewBufferString(raw))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return fail("expected a JSON object")
	}
	values := map[string]json.RawMessage{}
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return fail("invalid JSON object")
		}
		name, ok := key.(string)
		if !ok {
			return fail("invalid property name")
		}
		if _, exists := values[name]; exists {
			return fail("duplicate property " + name)
		}
		if name != "type" && name != "format" {
			return fail("unsupported property " + name)
		}
		var value json.RawMessage
		if decoder.Decode(&value) != nil {
			return fail("invalid property value")
		}
		values[name] = value
	}
	if _, err = decoder.Token(); err != nil {
		return fail("invalid JSON object")
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF {
		return fail("unexpected trailing JSON")
	}
	var types []string
	var single string
	if json.Unmarshal(values["type"], &single) == nil && single != "" {
		types = []string{single}
	} else if json.Unmarshal(values["type"], &types) != nil || len(types) == 0 {
		return fail("type must name one or more JSON scalars")
	}
	seen := map[string]bool{}
	for _, kind := range types {
		switch kind {
		case "string", "integer", "number", "boolean", "null":
		default:
			return fail("unsupported scalar type " + kind)
		}
		if seen[kind] {
			return fail("duplicate scalar type " + kind)
		}
		seen[kind] = true
	}
	if field.Type.Kind() == reflect.Pointer && !seen["null"] {
		types = append(types, "null")
	}
	result := map[string]any{}
	if len(types) == 1 {
		result["type"] = types[0]
	} else {
		result["type"] = types
	}
	if rawFormat, ok := values["format"]; ok {
		var format string
		if json.Unmarshal(rawFormat, &format) != nil || format == "" || !seen["string"] {
			return fail("format requires a string type and nonempty string value")
		}
		result["format"] = format
	}
	return result, true, nil
}
