package jsonschema_test

import (
	"github.com/stretchr/testify/require"
	"github.com/viant/mcp-protocol/schema/jsonschema"
	"reflect"
	"strings"
	"testing"
)

func TestFormattedEmbeddedSchemaDominance(t *testing.T) {
	type Embedded struct{ Id int }
	type Row struct {
		Embedded
		Id *int `json:"id"`
	}
	reflector := jsonschema.Reflector{FallbackName: func(f reflect.StructField, n string) string {
		if f.Tag.Get("json") != "" {
			return n
		}
		return strings.ToLower(n[:1]) + n[1:]
	}}
	result, err := reflector.Compile(jsonschema.Request{Type: reflect.TypeFor[Row]()})
	require.NoError(t, err)
	properties := result["properties"].(map[string]any)
	require.Len(t, properties, 1)
	require.Contains(t, properties, "id")
	reflector.ExcludeField = func(f reflect.StructField) bool { return f.Type.Kind() == reflect.Pointer }
	result, err = reflector.Compile(jsonschema.Request{Type: reflect.TypeFor[Row]()})
	require.NoError(t, err)
	require.Empty(t, result["properties"])
}

func TestFormattedSameDepthJSONTagDominance(t *testing.T) {
	type Row struct {
		Id    int
		Other string `json:"id"`
	}
	reflector := jsonschema.Reflector{FallbackName: func(f reflect.StructField, n string) string {
		if f.Tag.Get("json") != "" {
			return n
		}
		return strings.ToLower(n)
	}}
	result, err := reflector.Compile(jsonschema.Request{Type: reflect.TypeFor[Row]()})
	require.NoError(t, err)
	properties := result["properties"].(map[string]any)
	require.Len(t, properties, 1)
	require.Equal(t, "string", properties["id"].(map[string]any)["type"])
}
