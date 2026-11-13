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
