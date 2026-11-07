package jsonschema_test

import (
	"github.com/stretchr/testify/require"
	"github.com/viant/mcp-protocol/schema/jsonschema"
	"github.com/xeipuuv/gojsonschema"
	"reflect"
	"testing"
)

func TestNativeInputGuardSchema(t *testing.T) {
	type row struct {
		Value int `json:"value" format:"inputUnless=../Canonical" required:"true"`
	}
	type root struct {
		Canonical *int              `json:"canonical"`
		Rows      map[string][]*row `json:"rows"`
	}
	for _, native := range []bool{false, true} {
		schema, err := (jsonschema.Reflector{InputAliases: native}).Compile(jsonschema.Request{Type: reflect.TypeFor[root]()})
		require.NoError(t, err)
		validator, err := gojsonschema.NewSchema(gojsonschema.NewGoLoader(schema))
		require.NoError(t, err)
		result, err := validator.Validate(gojsonschema.NewStringLoader(`{"canonical":null,"rows":{"a":[{"value":"ignored"}]}}`))
		require.NoError(t, err)
		require.Equal(t, native, result.Valid(), "%v", result.Errors())
	}
	_, err := (jsonschema.Reflector{InputAliases: true}).Compile(jsonschema.Request{Type: reflect.TypeFor[row]()})
	require.Error(t, err)
}

func TestInputGuardOutputAndRecursiveContexts(t *testing.T) {
	type invalid struct {
		Value int `json:"value" format:"inputUnless=../Missing"`
	}
	_, err := (jsonschema.Reflector{InputAliases: true, Output: true}).Compile(jsonschema.Request{Type: reflect.TypeFor[invalid]()})
	require.NoError(t, err)
	type recursive struct {
		Canonical *int       `json:"canonical"`
		Value     int        `json:"value" format:"inputUnless=../Canonical"`
		Next      *recursive `json:"next"`
	}
	type root struct {
		Canonical *int       `json:"canonical"`
		Child     *recursive `json:"child"`
	}
	_, err = (jsonschema.Reflector{InputAliases: true}).Compile(jsonschema.Request{Type: reflect.TypeFor[root]()})
	require.NoError(t, err)
	_, err = (jsonschema.Reflector{InputAliases: true}).Compile(jsonschema.Request{Type: reflect.TypeFor[recursive]()})
	require.Error(t, err)
}
