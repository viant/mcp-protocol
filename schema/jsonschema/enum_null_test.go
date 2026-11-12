package jsonschema

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/viant/tagly/choice"
)

type explicitEnumNull struct {
	Code   *int                `json:"code,omitempty" choice:"0:ZERO" choice:"1:ONE" enumAccept:"literal" enumStyle:"literal" enumNull:"reject" format:"nullable=true"`
	Union  *choice.ScalarValue `json:"union,omitempty" choice:"1:ONE" enumAccept:"scalar" enumStyle:"scalar" enumNull:"reject" format:"nullable=true"`
	Legacy *int                `json:"legacy,omitempty" choice:"0:ZERO" choice:"1:ONE" enumAccept:"literal" enumStyle:"literal"`
}

func TestEnumNullRuleIsInputOnly(t *testing.T) {
	for _, output := range []bool{false, true} {
		schema, err := (Reflector{InputAliases: !output, Output: output}).Compile(Request{Type: reflect.TypeFor[explicitEnumNull]()})
		require.NoError(t, err)
		properties := schema["properties"].(map[string]any)
		code := properties["code"].(map[string]any)
		union := properties["union"].(map[string]any)
		legacy := properties["legacy"].(map[string]any)
		require.Equal(t, []string{"string", "null"}, legacy["type"])
		if output {
			require.Equal(t, []string{"string", "null"}, code["type"])
			require.Equal(t, []any{"ZERO", "ONE", nil}, code["enum"])
			require.NotContains(t, union, "not")
		} else {
			require.Equal(t, "string", code["type"])
			require.Equal(t, []any{"ZERO", "ONE"}, code["enum"])
			require.Equal(t, map[string]any{"type": "null"}, union["not"], "input policy wins over presentation nullable=true")
		}
		require.Empty(t, schema["required"], "null rejection must not make omission illegal")
	}
}
