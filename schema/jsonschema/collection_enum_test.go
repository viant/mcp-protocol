package jsonschema

import (
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestMCPCollectionEnumItems(t *testing.T) {
	type fields struct {
		Channels []string  `json:"channels" mcpEnum:"true" choice:"DISPLAY" choice:"VIDEO"`
		Optional *[]string `json:"optional" mcpEnum:"true" choice:"DISPLAY" choice:"VIDEO"`
		Scalar   string    `json:"scalar" mcpEnum:"true" choice:"1:ACTIVE" enumAccept:"scalar" enumStyle:"literal"`
		NoNull   []string  `json:"noNull" mcpEnum:"true" choice:"DISPLAY" format:"nullable=false"`
		Ordinary []string  `json:"ordinary"`
	}
	for _, output := range []bool{false, true} {
		document, err := (Reflector{Output: output}).Compile(Request{Type: reflect.TypeOf(fields{})})
		require.NoError(t, err)
		properties := document["properties"].(map[string]any)
		for _, name := range []string{"channels", "optional"} {
			property := properties[name].(map[string]any)
			require.Equal(t, []string{"array", "null"}, property["type"])
			require.NotContains(t, property, "enum")
			items := property["items"].(map[string]any)
			require.Equal(t, "string", items["type"])
			require.Equal(t, []any{"DISPLAY", "VIDEO"}, items["enum"])
		}
		require.Equal(t, "array", properties["noNull"].(map[string]any)["type"])
		require.Equal(t, "array", properties["ordinary"].(map[string]any)["type"])
	}
}
