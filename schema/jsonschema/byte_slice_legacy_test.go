package jsonschema

import (
	"reflect"
	"testing"
)

func TestPlainByteSliceReflectsLegacyIntegerArray(t *testing.T) {
	type uploadBody struct {
		Data []byte `json:"data"`
	}
	got, err := (Reflector{}).Compile(Request{Type: reflect.TypeFor[uploadBody]()})
	if err != nil {
		t.Fatal(err)
	}
	properties := got["properties"].(map[string]any)
	data := properties["data"].(map[string]any)
	if data["type"] != "array" {
		t.Fatalf("data schema type = %#v; want array", data["type"])
	}
	items := data["items"].(map[string]any)
	if items["type"] != "integer" {
		t.Fatalf("data item schema type = %#v; want integer", items["type"])
	}
	if _, ok := data["contentEncoding"]; ok {
		t.Fatalf("legacy byte array schema unexpectedly has contentEncoding: %#v", data)
	}
}
