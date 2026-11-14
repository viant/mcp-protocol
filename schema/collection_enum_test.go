package schema

import (
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestStructToPropertiesExplicitCollectionEnum(t *testing.T) {
	type input struct {
		Channels []string `json:"channels" mcpEnum:"true" enumAccept:"literal" choice:"DISPLAY" choice:"VIDEO"`
	}
	properties, _ := StructToProperties(reflect.TypeFor[input]())
	require.Equal(t, "array", properties["channels"]["type"])
	require.NotContains(t, properties["channels"], "enum")
	require.Equal(t, []any{"DISPLAY", "VIDEO"}, properties["channels"]["items"].(map[string]any)["enum"])
}

func TestStructToPropertiesNumericCollectionEnumUsesScalarElement(t *testing.T) {
	type fields struct {
		Codes []int `json:"codes" mcpEnum:"true" enumAccept:"literal" enumStyle:"literal" choice:"3:THREE" choice:"7:SEVEN"`
		Code  int   `json:"code" enumAccept:"literal" enumStyle:"literal" choice:"3:THREE" choice:"7:SEVEN"`
	}
	properties, _ := StructToProperties(reflect.TypeFor[fields]())
	require.NotContains(t, properties["codes"], "not")
	require.Equal(t, properties["code"], properties["codes"]["items"])
	require.NotContains(t, properties["codes"], "enum")
}
