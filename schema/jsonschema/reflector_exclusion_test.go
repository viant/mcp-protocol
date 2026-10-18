package jsonschema_test

import (
	"reflect"
	"testing"

	"github.com/viant/mcp-protocol/schema/jsonschema"
)

type hiddenJSONOwner struct {
	Leaked string `json:"leaked"`
}
type visibleJSONOwner struct {
	Shared string `json:"shared"`
}
type excludedJSONInput struct {
	*hiddenJSONOwner `mcp:"-"`
	visibleJSONOwner
	Shared string `json:"shared" mcp:"-"`
	Period string `json:"period" mcp:"-"`
	From   string `json:"from"`
}

func excludeMCPField(field reflect.StructField) bool { return field.Tag.Get("mcp") == "-" }

func TestReflectorOwnerExclusionPreservesJSONDefaultsAndDominance(t *testing.T) {
	ordinary, err := (jsonschema.Reflector{}).Compile(jsonschema.Request{Type: reflect.TypeFor[excludedJSONInput]()})
	if err != nil {
		t.Fatal(err)
	}
	if len(ordinary["properties"].(map[string]any)) != 4 {
		t.Fatal("default JSON visibility changed")
	}
	filtered, err := (jsonschema.Reflector{ExcludeField: excludeMCPField}).Compile(jsonschema.Request{Type: reflect.TypeFor[excludedJSONInput]()})
	if err != nil {
		t.Fatal(err)
	}
	fields := filtered["properties"].(map[string]any)
	if len(fields) != 1 || fields["from"] == nil {
		t.Fatalf("excluded owners or shadowed fields leaked: %v", fields)
	}
}

func TestReflectorExclusionRunsBeforeNestedSchemaAndAnnotation(t *testing.T) {
	type filters struct {
		Opaque chan int `json:"opaque" mcp:"-"`
		From   string   `json:"from"`
	}
	type input struct {
		Filters *filters           `json:"filters"`
		Rows    []filters          `json:"rows"`
		ByName  map[string]filters `json:"byName"`
	}
	annotatedHidden := false
	result, err := (jsonschema.Reflector{ExcludeField: excludeMCPField, Annotate: func(_ string, f reflect.StructField) (string, any) {
		annotatedHidden = annotatedHidden || f.Name == "Opaque"
		return "", nil
	}}).Compile(jsonschema.Request{Type: reflect.TypeFor[input]()})
	if err != nil || annotatedHidden {
		t.Fatalf("hidden child was compiled or annotated: %v", err)
	}
	properties := result["properties"].(map[string]any)
	for _, name := range []string{"filters", "rows", "byName"} {
		node := properties[name].(map[string]any)
		if name == "rows" {
			node = node["items"].(map[string]any)
		}
		if name == "byName" {
			node = node["additionalProperties"].(map[string]any)
		}
		fields := node["properties"].(map[string]any)
		if len(fields) != 1 || fields["from"] == nil {
			t.Fatalf("nested %s exclusion ignored: %v", name, fields)
		}
	}
}
