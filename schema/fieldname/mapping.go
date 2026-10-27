package fieldname

import (
	"encoding"
	"encoding/json"
	"fmt"
	xshape "github.com/viant/x/shape"
	"reflect"
)

// Fields uses the same native JSON winner selection as the modern reflector.
// It supplies exact original names/indices; it never invents a second shape.
func Fields(t reflect.Type, visibility ...Visibility) ([]Entry, error) {
	if t == nil {
		return nil, fmt.Errorf("MCP field type is required")
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	fields, err := xshape.Linked(t).JSONFields()
	if err != nil {
		return nil, err
	}
	var out []Entry
	for _, f := range fields {
		if selectedVisibility(visibility).Excluded(t, f.Field.Index) {
			continue
		}

		if f.Field.Tag.Get("setMarker") == "true" {
			continue
		}
		n, e := selectedVisibility(visibility).Resolve(f.Field.StructField(), f.Name)
		if e != nil {
			return nil, e
		}
		n.Hidden = false // Visibility is selected by the schema owner before naming.
		out = append(out, Entry{Field: f.Field.StructField(), JSONName: f.Name, Name: n})
	}
	if err := Validate(out); err != nil {
		return nil, err
	}
	return out, nil
}
func CustomJSON(t reflect.Type) bool {
	if t == nil {
		return false
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	u := reflect.TypeFor[json.Unmarshaler]()
	text := reflect.TypeFor[encoding.TextUnmarshaler]()
	return t.Implements(u) || reflect.PointerTo(t).Implements(u) || t.Implements(text) || reflect.PointerTo(t).Implements(text)
}

// Normalize only translates explicitly authored MCP names/aliases into existing
// JSON member names in ordinary typed bodies. JSON decoding still validates all
// values. Unknown keys, absence, null and empty objects/arrays are retained.
// Supplying more than one spelling is always an error, even for equal values.
func Normalize(t reflect.Type, value any, visibility ...Visibility) (any, error) {
	if t == nil {
		return nil, fmt.Errorf("MCP field type is required")
	}
	if value == nil {
		return nil, nil
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if CustomJSON(t) {
		explicit := selectedVisibility(visibility)
		explicit.FallbackName = nil
		yes, e := HasNames(t, explicit)
		if e != nil {
			return nil, e
		}
		if yes {
			return nil, fmt.Errorf("MCP names on custom JSON type %s require decoding authority", t)
		}
		return value, nil
	}

	if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		a, ok := value.([]any)
		if !ok {
			return value, nil
		}
		out := make([]any, len(a))
		for i, v := range a {
			n, e := Normalize(t.Elem(), v, visibility...)
			if e != nil {
				return nil, e
			}
			out[i] = n
		}
		return out, nil
	}
	if t.Kind() == reflect.Map {
		m, ok := value.(map[string]any)
		if !ok {
			return value, nil
		}
		out := make(map[string]any, len(m))
		for key, v := range m {
			n, e := Normalize(t.Elem(), v, visibility...)
			if e != nil {
				return nil, e
			}
			out[key] = n
		}
		return out, nil
	}
	if t.Kind() != reflect.Struct {
		return value, nil
	}
	entries, e := Fields(t, visibility...)
	if e != nil {
		return nil, e
	}

	m, ok := value.(map[string]any)
	if !ok {
		return value, nil
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	for _, f := range entries {

		spellings := []string{f.JSONName}
		if f.Name.Public != f.JSONName {
			spellings = append(spellings, f.Name.Public)
		}
		spellings = append(spellings, f.Name.Aliases...)
		seen := map[string]bool{}
		key := ""
		for _, name := range spellings {
			if seen[name] {
				continue
			}
			seen[name] = true
			if _, supplied := m[name]; supplied {
				if key != "" {
					return nil, fmt.Errorf("conflicting MCP names %q and %q for JSON field %q", key, name, f.JSONName)
				}
				key = name
			}
		}
		if key == "" {
			continue
		}
		v, e := Normalize(f.Field.Type, m[key], visibility...)
		if e != nil {
			return nil, e
		}
		delete(out, key)
		out[f.JSONName] = v
	}
	return out, nil
}

// HasNames checks the selected ordinary subtree without interpreting custom
// decoder behavior. A custom owner must provide separate decoding authority if
// any descendant would acquire an MCP spelling.
func HasNames(t reflect.Type, visibility ...Visibility) (bool, error) {
	if t == nil {
		return false, fmt.Errorf("MCP field type is required")
	}
	type visitKey struct {
		typ       reflect.Type
		formatted bool
	}
	seen := map[visitKey]bool{}
	var visit func(reflect.Type, Visibility) (bool, error)
	visit = func(t reflect.Type, policy Visibility) (bool, error) {
		if CustomJSON(t) {
			policy.FallbackName = nil
		}
		key := visitKey{t, policy.FallbackName != nil}
		if seen[key] {
			return false, nil
		}
		seen[key] = true
		switch t.Kind() {
		case reflect.Pointer, reflect.Array, reflect.Slice, reflect.Map:
			return visit(t.Elem(), policy)
		}
		if t.Kind() != reflect.Struct {
			return false, nil
		}

		fields, e := Fields(t, policy)
		if e != nil {
			return false, e
		}
		for _, f := range fields {
			if f.Name.Explicit || len(f.Name.Aliases) > 0 || f.Name.Public != f.JSONName {
				return true, nil
			}
			yes, e := visit(f.Field.Type, policy)
			if e != nil || yes {
				return yes, e
			}
		}
		return false, nil
	}
	return visit(t, selectedVisibility(visibility))
}
