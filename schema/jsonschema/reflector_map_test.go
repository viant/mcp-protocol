package jsonschema_test

import (
	"encoding/json"
	"github.com/viant/mcp-protocol/schema/jsonschema"
	"reflect"
	"testing"
)

type recordKey int64

func TestReflectorIntegerMapKeysPreserveJSONObjects(t *testing.T) {
	for _, value := range []any{map[int]string{-2: "v"}, map[int8]string{1: "v"}, map[int16]string{1: "v"}, map[int32]string{1: "v"}, map[int64]string{1: "v"}, map[uint]string{1: "v"}, map[uint8]string{1: "v"}, map[uint16]string{1: "v"}, map[uint32]string{1: "v"}, map[uint64]string{1: "v"}, map[uintptr]string{1: "v"}, map[recordKey]string{1: "v"}} {
		t.Run(reflect.TypeOf(value).String(), func(t *testing.T) {
			encoded, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			var object map[string]string
			if err = json.Unmarshal(encoded, &object); err != nil || len(object) != 1 {
				t.Fatal(string(encoded), err)
			}
			shape, err := (jsonschema.Reflector{}).Compile(jsonschema.Request{Type: reflect.TypeOf(value)})
			if err != nil {
				t.Fatal(err)
			}
			if shape["type"] != "object" || shape["additionalProperties"].(map[string]any)["type"] != "string" {
				t.Fatal(shape)
			}
		})
	}
	for _, typ := range []reflect.Type{reflect.TypeFor[map[bool]string](), reflect.TypeFor[map[float64]string]()} {
		if _, err := (jsonschema.Reflector{}).Compile(jsonschema.Request{Type: typ}); err == nil {
			t.Fatal("unsupported JSON key accepted", typ)
		}
	}
}
