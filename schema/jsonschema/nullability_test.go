package jsonschema

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

type nullableMetadataRecord struct {
	Forced  *bool                  `json:"forced,omitempty" required:"true" format:"nullable=false"`
	Preview *bool                  `json:"preview,omitempty" format:"nullable=false"`
	Default *bool                  `json:"default,omitempty"`
	Scalar  int                    `json:"scalar,omitempty" format:"nullable=true"`
	Code    *int                   `json:"code,omitempty" choice:"1:ONE" choice:"2:TWO" enumAccept:"literal" enumStyle:"literal" format:"nullable=false"`
	Values  *[]*int                `json:"values,omitempty" format:"nullable=false"`
	Lookup  *map[string]*int       `json:"lookup,omitempty" format:"nullable=false"`
	Nested  *nullableMetadataChild `json:"nested,omitempty" format:"nullable=false"`
}
type nullableMetadataChild struct {
	Value *bool `json:"value,omitempty"`
}

func TestAuthoredFieldNullability(t *testing.T) {
	for _, output := range []bool{false, true} {
		t.Run(map[bool]string{false: "input", true: "output"}[output], func(t *testing.T) {
			compiled, err := (Reflector{InputAliases: !output, Output: output}).Compile(Request{Type: reflect.TypeFor[nullableMetadataRecord]()})
			require.NoError(t, err)
			fields := compiled["properties"].(map[string]any)
			field := func(name string) map[string]any { return fields[name].(map[string]any) }
			require.Equal(t, "boolean", field("preview")["type"])
			require.Equal(t, []string{"boolean", "null"}, field("default")["type"], "untagged pointers retain existing behavior")
			require.Equal(t, []string{"integer", "null"}, field("scalar")["type"])
			require.Equal(t, "string", field("code")["type"])
			require.Equal(t, []any{"ONE", "TWO"}, field("code")["enum"])
			require.Equal(t, "array", field("values")["type"])
			require.Equal(t, "integer", field("values")["items"].(map[string]any)["type"], "existing slice-item policy remains unchanged")
			require.Equal(t, "object", field("lookup")["type"])
			require.Equal(t, []string{"integer", "null"}, field("lookup")["additionalProperties"].(map[string]any)["type"])
			require.Equal(t, "object", field("nested")["type"])
			child := field("nested")["properties"].(map[string]any)["value"].(map[string]any)
			require.Equal(t, []string{"boolean", "null"}, child["type"], "parent metadata cannot change child nullability")
			require.Equal(t, []string{"forced"}, compiled["required"], "nonnull does not change presence requirements")
		})
	}
}

func TestNullabilityPreservesSharedSchemas(t *testing.T) {
	original := map[string]any{"type": []string{"string", "null"}, "enum": []any{"ONE", nil}, "description": "unchanged"}
	nonnull := fieldNullability(original, false)
	require.Equal(t, "string", nonnull["type"])
	require.Equal(t, []any{"ONE"}, nonnull["enum"])
	require.Equal(t, "unchanged", nonnull["description"])
	require.Equal(t, []string{"string", "null"}, original["type"])
	require.Equal(t, []any{"ONE", nil}, original["enum"])
	for _, tc := range []struct {
		key   string
		value any
	}{
		{"$ref", "#/$defs/record"},
		{"anyOf", []any{map[string]any{"type": "string"}, map[string]any{"type": "null"}}},
		{"allOf", []any{map[string]any{"type": []string{"boolean", "null"}}}},
		{"oneOf", []any{map[string]any{"type": "integer"}, map[string]any{"type": "null"}}},
		{"not", map[string]any{"type": "integer"}},
	} {
		key := tc.key
		t.Run(key, func(t *testing.T) {
			schema := map[string]any{key: tc.value}
			nullable := fieldNullability(schema, true)
			require.Equal(t, []any{schema, map[string]any{"type": "null"}}, nullable["anyOf"])
			nonnull := fieldNullability(schema, false)
			if key == "$ref" || key == "not" {
				require.Equal(t, []any{schema, map[string]any{"not": map[string]any{"type": "null"}}}, nonnull["allOf"])
			} else {
				require.Equal(t, map[string]any{"type": "null"}, nonnull["not"])
			}
			require.Len(t, schema, 1)
		})
	}
}
