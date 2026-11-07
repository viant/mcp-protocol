package jsonschema_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/viant/mcp-protocol/schema/jsonschema"
	"github.com/xeipuuv/gojsonschema"
)

func TestNativeInputAliasAuthority(t *testing.T) {
	type body struct {
		Name string `json:"name" required:"true" format:"inputNames=oldName"`
	}
	for _, tc := range []struct{ enabled, output, alias bool }{{false, false, false}, {true, false, true}, {true, true, false}} {
		schema, err := (jsonschema.Reflector{InputAliases: tc.enabled, Output: tc.output}).Compile(jsonschema.Request{Type: reflect.TypeFor[body]()})
		require.NoError(t, err)
		_, exists := schema["properties"].(map[string]any)["oldName"]
		require.Equal(t, tc.alias, exists)
		for _, value := range []struct {
			body  map[string]any
			valid bool
		}{
			{map[string]any{"name": "ok"}, true}, {map[string]any{"oldName": "ok"}, tc.alias},
			{map[string]any{}, false}, {map[string]any{"oldName": 3}, false},
		} {
			result, err := gojsonschema.Validate(gojsonschema.NewGoLoader(schema), gojsonschema.NewGoLoader(value.body))
			require.NoError(t, err)
			require.Equal(t, value.valid, result.Valid(), "%v %v", value.body, result.Errors())
		}
	}
}

func TestNativeInputAliasPolicyConservativeAdmission(t *testing.T) {
	type constraint struct {
		Limit int `json:"limit"`
	}
	type body struct {
		Frequency *constraint `json:"frequencyCapping,omitempty" format:"inputNames=frequencyCap,inputNamesPriority=canonical,inputNamesKinds=object|null,inputNamesCase=fold"`
	}
	schema, err := (jsonschema.Reflector{InputAliases: true, RequireField: func(_ string, field reflect.StructField) bool { return field.Name == "Frequency" }}).Compile(jsonschema.Request{Type: reflect.TypeFor[body]()})
	require.NoError(t, err)
	for _, body := range []map[string]any{
		{"frequencyCap": 5},
		{"frequencyCapping": map[string]any{"limit": 2}, "frequencyCap": map[string]any{"limit": "ignored"}},
		{"FREQUENCYCAPPING": map[string]any{"limit": 2}, "frequencyCap": map[string]any{"limit": "ignored"}},
	} {
		result, err := gojsonschema.Validate(gojsonschema.NewGoLoader(schema), gojsonschema.NewGoLoader(body))
		require.NoError(t, err)
		require.True(t, result.Valid(), "%v", result.Errors())
	}
	require.Contains(t, schema["properties"].(map[string]any)["frequencyCap"].(map[string]any)["description"], "runtime")
	require.Empty(t, schema["required"], "folded required presence is enforced by runtime, not an exact-only discovery constraint")
}

func TestNativeInputAliasesNestedKindsAndVisibility(t *testing.T) {
	type constraint struct {
		Limit int `json:"limit"`
	}
	type leaf struct {
		Value  *constraint `json:"value" required:"true" format:"inputNames=oldValue,inputNamesKinds=object"`
		Hidden string      `json:"-" format:"inputNames=secret"`
	}
	type envelope struct {
		Items   []*leaf          `json:"items"`
		Mapping map[string]*leaf `json:"mapping"`
	}
	schema, err := (jsonschema.Reflector{InputAliases: true}).Compile(jsonschema.Request{Type: reflect.TypeFor[envelope]()})
	require.NoError(t, err)
	for _, tc := range []struct {
		value any
		valid bool
	}{
		{map[string]any{"limit": 2}, true}, {3, false}, {map[string]any{"limit": "bad"}, false},
	} {
		entry := map[string]any{"oldValue": tc.value}
		body := map[string]any{"items": []any{entry}, "mapping": map[string]any{"entry": entry}}
		result, err := gojsonschema.Validate(gojsonschema.NewGoLoader(schema), gojsonschema.NewGoLoader(body))
		require.NoError(t, err)
		require.Equal(t, tc.valid, result.Valid(), "%v", result.Errors())
	}
	item := schema["properties"].(map[string]any)["items"].(map[string]any)["items"].(map[string]any)
	require.NotContains(t, item["properties"], "secret")
}

func TestNativeInputAliasExclusionAndCollision(t *testing.T) {
	type excluded struct {
		Value string `json:"value" required:"true" format:"inputNames=oldValue|Ignored,inputNamesExclude=Ignored"`
	}
	schema, err := (jsonschema.Reflector{InputAliases: true}).Compile(jsonschema.Request{Type: reflect.TypeFor[excluded]()})
	require.NoError(t, err)
	require.NotContains(t, schema["properties"], "Ignored")
	result, err := gojsonschema.Validate(gojsonschema.NewGoLoader(schema), gojsonschema.NewGoLoader(map[string]any{"Ignored": "no"}))
	require.NoError(t, err)
	require.False(t, result.Valid())
	type collision struct {
		Name string `json:"name" format:"inputNames=id"`
		ID   int    `json:"id"`
	}
	_, err = (jsonschema.Reflector{InputAliases: true}).Compile(jsonschema.Request{Type: reflect.TypeFor[collision]()})
	require.ErrorContains(t, err, "conflicts with canonical")
}

func TestOwnerRequirednessPrecedesAliasProjection(t *testing.T) {
	type child struct {
		Name   string `json:"name,omitempty" format:"inputNames=oldName"`
		Code   string `json:"code,omitempty"`
		Hidden string `json:"-"`
	}
	type body struct {
		Child *child `json:"child"`
	}
	var paths []string
	schema, err := (jsonschema.Reflector{InputAliases: true, RequireField: func(path string, field reflect.StructField) bool {
		paths = append(paths, path)
		return path == "Body.Child.Name" || path == "Body.Child.Code"
	}}).Compile(jsonschema.Request{Type: reflect.TypeFor[body](), Path: "Body"})
	require.NoError(t, err)
	require.NotContains(t, paths, "Body.Child.Hidden")
	for _, tc := range []struct {
		child map[string]any
		valid bool
	}{
		{map[string]any{"oldName": "a", "code": "b"}, true},
		{map[string]any{"name": "a", "code": "b"}, true},
		{map[string]any{"code": "b"}, false},
		{map[string]any{"oldName": "a"}, false},
	} {
		result, err := gojsonschema.Validate(gojsonschema.NewGoLoader(schema), gojsonschema.NewGoLoader(map[string]any{"child": tc.child}))
		require.NoError(t, err)
		require.Equal(t, tc.valid, result.Valid(), "%v", result.Errors())
	}
}
