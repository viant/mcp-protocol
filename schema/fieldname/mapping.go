package fieldname

import (
	"encoding"
	"encoding/json"
	"fmt"
	xshape "github.com/viant/x/shape"
	"reflect"
	"strings"
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
	policy := selectedVisibility(visibility)
	policy.Representation = metadataForType(policy.Representation, t)
	var fields []xshape.JSONField
	var err error
	if policy.Representation != nil {
		if policy.Representation.Original() != t {
			return nil, fmt.Errorf("MCP metadata type mismatch")
		}
		fields, err = policy.Representation.JSONFields()
	} else {
		fields, err = xshape.Linked(t).JSONFields()
	}
	if err != nil {
		return nil, err
	}
	var out []Entry
	for _, f := range fields {
		if f.Field.Tag.Get("setMarker") == "true" {
			continue
		}
		n, e := policy.Resolve(f.Field.StructField(), f.Name)
		if e != nil {
			return nil, e
		}
		n.Hidden = false // Visibility is selected by the schema owner before naming.
		var metadata *xshape.Representation
		if policy.Representation != nil {
			metadata = policy.Representation.FieldShapeByIndex(f.Field.Index)
		}
		out = append(out, Entry{Representation: metadata, Field: f.Field.StructField(), JSONName: f.Name, Name: n})
	}
	// Case conversion can merge otherwise distinct native JSON names across
	// promoted fields. Select the unique shallow winner before visibility so
	// excluding that winner cannot expose the shadowed member.
	if policy.FallbackName != nil {
		out = formattedWinners(out)
	}
	visible := out[:0]
	for _, entry := range out {
		if !policy.Excluded(t, entry.Field.Index) {
			visible = append(visible, entry)
		}
	}
	out = visible
	if err := Validate(out); err != nil {
		return nil, err
	}
	return out, nil
}

// formattedWinners resolves only collisions introduced by implicit formatting.
// Native JSON tag dominance breaks same-depth ties. Authored MCP names,
// aliases and remaining ambiguities stay Validate's concern.
func formattedWinners(entries []Entry) []Entry {
	groups := map[string][]int{}
	for i, entry := range entries {
		groups[entry.Name.Public] = append(groups[entry.Name.Public], i)
	}
	shadowed := map[int]bool{}
	for _, group := range groups {
		if len(group) < 2 {
			continue
		}
		winner, depth, unique := -1, int(^uint(0)>>1), false
		formatted, explicit := false, false
		for _, i := range group {
			entry := entries[i]
			formatted = formatted || entry.Name.Public != entry.JSONName
			explicit = explicit || entry.Name.Explicit
			switch n := len(entry.Field.Index); {
			case n < depth:
				winner, depth, unique = i, n, true
			case n == depth:
				unique = false
			}
		}
		if !unique {
			tagged := 0
			for _, i := range group {
				entry := entries[i]
				name, _, _ := strings.Cut(entry.Field.Tag.Get("json"), ",")
				if len(entry.Field.Index) == depth && name != "" && name != "-" {
					winner = i
					tagged++
				}
			}
			unique = tagged == 1
		}
		if !formatted || explicit || !unique {
			continue
		}
		for _, i := range group {
			if i != winner {
				shadowed[i] = true
			}
		}
	}
	result := entries[:0]
	for i, entry := range entries {
		if !shadowed[i] {
			result = append(result, entry)
		}
	}
	return result
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
	policy := selectedVisibility(visibility)
	policy.Representation = metadataForType(policy.Representation, t)
	if CustomJSON(t) {
		explicit := policy
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
			n, e := Normalize(t.Elem(), v, elementVisibility(policy))
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
			n, e := Normalize(t.Elem(), v, elementVisibility(policy))
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
	entries, e := Fields(t, policy)
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
		v, e := Normalize(f.Field.Type, m[key], childVisibility(policy, f.Representation))
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
		metadata  *xshape.Representation
	}
	seen := map[visitKey]bool{}
	var visit func(reflect.Type, Visibility) (bool, error)
	visit = func(t reflect.Type, policy Visibility) (bool, error) {
		if CustomJSON(t) {
			policy.FallbackName = nil
		}
		key := visitKey{t, policy.FallbackName != nil, policy.Representation}
		if seen[key] {
			return false, nil
		}
		seen[key] = true
		switch t.Kind() {
		case reflect.Pointer, reflect.Array, reflect.Slice, reflect.Map:
			return visit(t.Elem(), elementVisibility(policy))
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
			yes, e := visit(f.Field.Type, childVisibility(policy, f.Representation))
			if e != nil || yes {
				return yes, e
			}
		}
		return false, nil
	}
	return visit(t, selectedVisibility(visibility))
}
