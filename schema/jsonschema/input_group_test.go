package jsonschema_test

import (
	"github.com/stretchr/testify/require"
	"github.com/viant/mcp-protocol/schema/jsonschema"
	"github.com/xeipuuv/gojsonschema"
	"reflect"
	"testing"
)

func TestNativeInputGroupsConservativeSchema(t *testing.T) {
	type row struct {
		Canonical *int16 `json:"day_of_week" required:"true" format:"inputGroup=weekday,inputPriority=0,inputNames=dayOfWeek"`
		Upper     *int   `json:"Weekday" format:"inputGroup=weekday,inputPriority=1,inputCase=exact"`
		Lower     *int16 `json:"weekday" format:"inputGroup=weekday,inputPriority=2,inputCase=exact"`
		Hidden    *int   `json:"-" format:"inputGroup=weekday,inputPriority=2"`
	}
	for _, native := range []bool{false, true} {
		for _, output := range []bool{false, true} {
			schema, err := (jsonschema.Reflector{InputAliases: native, Output: output}).Compile(jsonschema.Request{Type: reflect.TypeFor[row]()})
			require.NoError(t, err)
			validator, err := gojsonschema.NewSchema(gojsonschema.NewGoLoader(schema))
			require.NoError(t, err)
			result, err := validator.Validate(gojsonschema.NewStringLoader(`{"day_of_week":"losing","Weekday":{},"weekday":null}`))
			require.NoError(t, err)
			require.Equal(t, native && !output, result.Valid(), "%v", result.Errors())
			if native && !output {
				require.Empty(t, schema["required"])
				for _, name := range []string{"day_of_week", "dayOfWeek", "Weekday", "weekday"} {
					property := schema["properties"].(map[string]any)[name].(map[string]any)
					require.Contains(t, property["description"], "runtime")
				}
			}
		}
	}
}
func TestNativeInputGroupsValidateConfiguration(t *testing.T) {
	type invalid struct {
		A int `json:"a" format:"inputGroup=x,inputPriority=0"`
		B int `json:"b" format:"inputGroup=x,inputPriority=0"`
	}
	_, err := (jsonschema.Reflector{InputAliases: true}).Compile(jsonschema.Request{Type: reflect.TypeFor[invalid]()})
	require.Error(t, err)
}
