package jsonschema

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOutputAliasesOpenAndStrictDomains(t *testing.T) {
	type values struct {
		Open     string  `json:"open" choice:"1:Canonical" enumAccept:"literal" enumStyle:"literal" enumOutputAlias:"1:stored"`
		Optional *string `json:"optional" choice:"1:Canonical" enumAccept:"literal" enumStyle:"literal" enumOutputAlias:"1:stored"`
		Strict   string  `json:"strict" choice:"1:Canonical" enumAccept:"literal" enumStyle:"literal" enumOutputAlias:"1:stored" enumOutputUnknown:"reject"`
	}
	for _, output := range []bool{false, true} {
		doc, err := (Reflector{Output: output}).Compile(Request{Type: reflect.TypeFor[values]()})
		require.NoError(t, err)
		properties := doc["properties"].(map[string]any)
		for _, name := range []string{"open", "optional", "strict"} {
			field := properties[name].(map[string]any)
			if output && name != "strict" {
				require.NotContains(t, field, "enum", name)
			} else {
				require.Contains(t, field, "enum", name)
				require.NotContains(t, field["enum"], "stored", "output aliases never become admitted inputs")
			}
		}
		require.Equal(t, "string", properties["open"].(map[string]any)["type"])
		require.Equal(t, []string{"string", "null"}, properties["optional"].(map[string]any)["type"])
	}
}
