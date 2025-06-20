package schema

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildJSONSchema_Cycles(t *testing.T) {
	type A struct{ Next *A }
	type B struct{ Others []B }
	type C struct{ M map[string]C }

	tests := []struct {
		name     string
		typ      reflect.Type
		expected map[string]interface{}
	}{
		{
			name: "pointer cycle",
			typ:  reflect.TypeOf(A{}),
			expected: map[string]interface{}{
				"type": "object",
				"properties": ToolInputSchemaProperties{
					"Next": map[string]interface{}{
						"type":     "object",
						"nullable": true,
					},
				},
			},
		},
		{
			name: "slice cycle",
			typ:  reflect.TypeOf(B{}),
			expected: map[string]interface{}{
				"type": "object",
				"properties": ToolInputSchemaProperties{
					"Others": map[string]interface{}{
						"type":  "array",
						"items": map[string]interface{}{"type": "object"},
					},
				},
				"required": []string{"Others"},
			},
		},
		{
			name: "map cycle",
			typ:  reflect.TypeOf(C{}),
			expected: map[string]interface{}{
				"type": "object",
				"properties": ToolInputSchemaProperties{
					"M": map[string]interface{}{
						"type":                 "object",
						"additionalProperties": map[string]interface{}{"type": "object"},
					},
				},
				"required": []string{"M"},
			},
		},
	}

	for _, tc := range tests {
		actual := typeSchema(tc.typ)
		assert.EqualValues(t, tc.expected, actual, tc.name)
	}
}

func TestBuildJSONSchema_Interface(t *testing.T) {
	type S struct {
		Any  interface{}
		Meta map[string]interface{}
	}

	expected := map[string]interface{}{
		"type": "object",
		"properties": ToolInputSchemaProperties{
			"Any": map[string]interface{}{},
			"Meta": map[string]interface{}{
				"type":                 "object",
				"additionalProperties": true,
			},
		},
		"required": []string{"Any", "Meta"},
	}

	actual := typeSchema(reflect.TypeOf(S{}))
	assert.EqualValues(t, expected, actual)
}
