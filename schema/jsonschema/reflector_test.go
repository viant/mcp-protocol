package jsonschema_test

import (
	"encoding/json"
	"github.com/viant/mcp-protocol/schema/jsonschema"
	"reflect"
	"strings"
	"testing"
)

type Embedded struct {
	ID int `json:"identifier" desc:"authored identifier"`
}
type Node struct {
	Embedded
	Name   string `json:"renamed"`
	Next   *Node  `json:"next,omitempty"`
	Hidden string `json:"-"`
}

func TestRecursiveEmbeddedAnnotationProjection(t *testing.T) {
	reflector := jsonschema.Reflector{Annotate: func(path string, field reflect.StructField) (string, any) {
		if value := field.Tag.Get("desc"); value != "" {
			return value, nil
		}
		return path, "example"
	}}
	result, err := reflector.Compile(jsonschema.Request{Type: reflect.TypeFor[Node](), Path: "Payload", Property: "payload/~"})
	if err != nil {
		t.Fatal(err)
	}
	properties := result["properties"].(map[string]any)
	if _, ok := properties["Embedded"]; ok {
		t.Fatal("embedded holder leaked")
	}
	if _, ok := properties["Hidden"]; ok {
		t.Fatal("hidden field leaked")
	}
	if properties["identifier"].(map[string]any)["description"] != "authored identifier" {
		t.Fatal(properties)
	}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "#/properties/payload~1~0/$defs/") {
		t.Fatal(string(data))
	}
	// Repeated compilations do not retain mutable schema or per-call annotations.
	properties["identifier"].(map[string]any)["type"] = "changed"
	again, err := reflector.Compile(jsonschema.Request{Type: reflect.TypeFor[Node](), Path: "Payload", Property: "payload/~"})
	if err != nil {
		t.Fatal(err)
	}
	if again["properties"].(map[string]any)["identifier"].(map[string]any)["type"] != "integer" {
		t.Fatal(again)
	}
}
