package jsonschema

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBooleanEnumWireSchemas(t *testing.T) {
	type flags struct {
		Flag       int  `json:"flag" choice:"0:false" choice:"1:true" enumAccept:"boolean" enumStyle:"boolean"`
		Optional   *int `json:"optional" choice:"0:false" choice:"1:true" enumAccept:"boolean" enumStyle:"boolean"`
		RejectNull *int `json:"rejectNull" choice:"0:false" choice:"1:true" enumAccept:"boolean" enumStyle:"boolean" enumNull:"reject"`
		InputOnly  int  `json:"inputOnly" choice:"0:false" choice:"1:true" enumAccept:"boolean" enumStyle:"ordinal"`
		OutputOnly int  `json:"outputOnly" choice:"0:false" choice:"1:true" enumAccept:"ordinal" enumStyle:"boolean"`
	}
	for _, output := range []bool{false, true} {
		document, err := (Reflector{Output: output}).Compile(Request{Type: reflect.TypeFor[flags]()})
		require.NoError(t, err)
		properties := document["properties"].(map[string]any)
		flag := properties["flag"].(map[string]any)
		require.Equal(t, "boolean", flag["type"])
		require.Equal(t, []any{false, true}, flag["enum"])
		optional := properties["optional"].(map[string]any)
		require.Equal(t, []string{"boolean", "null"}, optional["type"])
		require.Equal(t, []any{false, true, nil}, optional["enum"])
		rejected := properties["rejectNull"].(map[string]any)
		if output {
			// enumNull owns input admission; a nil stored pointer remains null.
			require.Equal(t, []string{"boolean", "null"}, rejected["type"])
			require.Equal(t, []any{false, true, nil}, rejected["enum"])
		} else {
			require.Equal(t, "boolean", rejected["type"])
			require.Equal(t, []any{false, true}, rejected["enum"])
		}
		for name, boolean := range map[string]bool{"inputOnly": !output, "outputOnly": output} {
			property := properties[name].(map[string]any)
			if boolean {
				require.Equal(t, "boolean", property["type"])
				require.Equal(t, []any{false, true}, property["enum"])
			} else {
				require.Equal(t, "integer", property["type"])
				require.Equal(t, []any{int64(0), int64(1)}, property["enum"])
			}
		}
	}
}
