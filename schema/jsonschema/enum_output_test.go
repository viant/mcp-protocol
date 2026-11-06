package jsonschema

import (
	"reflect"
	"testing"
)

func TestEnumOutputUsesStyleIndependentlyOfInput(t *testing.T) {
	type value struct {
		Mode int `json:"mode" choice:"0:draft" choice:"10:active" enumAccept:"literal" enumStyle:"ordinal"`
	}
	for _, output := range []bool{false, true} {
		doc, err := (Reflector{Output: output}).Compile(Request{Type: reflect.TypeFor[value]()})
		if err != nil {
			t.Fatal(err)
		}
		field := doc["properties"].(map[string]any)["mode"].(map[string]any)
		want := "string"
		var values any = []any{"draft", "active"}
		if output {
			want = "integer"
			values = []any{int64(0), int64(10)}
		}
		if field["type"] != want || !reflect.DeepEqual(field["enum"], values) {
			t.Fatalf("output=%t field=%+v", output, field)
		}
	}
}
