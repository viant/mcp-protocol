package jsonschema

import (
	"encoding/json"
	"github.com/viant/tagly/choice"
	"github.com/viant/x/shape"
	"reflect"
	"strings"
	"testing"
)

func TestScalarUnionReflector(t *testing.T) {
	type row struct {
		Fee choice.ScalarValue `json:"fee" choice:"4237:AD_SERVING" choice:"4241:AD_SERVING" enumAccept:"scalar" enumStyle:"scalar" enumInteger:"passthrough" enumUnknownString:"preserve"`
	}
	for _, output := range []bool{false, true} {
		doc, err := (Reflector{Output: output}).Compile(Request{Type: reflect.TypeFor[row]()})
		if err != nil {
			t.Fatal(err)
		}
		field := doc["properties"].(map[string]any)["fee"].(map[string]any)
		if _, ok := field["type"]; ok {
			t.Fatalf("stale scalar type: %v", field)
		}
		if _, ok := field["$ref"]; ok {
			t.Fatalf("storage ref: %v", field)
		}
		branches, ok := field["anyOf"].([]any)
		if !ok || len(branches) != 3 {
			t.Fatalf("union: %v", field)
		}
		data, err := json.Marshal(branches[0])
		if err != nil {
			t.Fatal(err)
		}
		var integer map[string]json.RawMessage
		if err = json.Unmarshal(data, &integer); err != nil {
			t.Fatal(err)
		}
		if string(integer["maximum"]) != "9223372036854775807" || string(integer["minimum"]) != "-9223372036854775808" {
			t.Fatalf("bounds %s", data)
		}
	}
}

func TestScalarUnionNestedAliasesAndAnnotations(t *testing.T) {
	type fee struct {
		Value choice.ScalarValue `json:"value" choice:"4237:AD_SERVING" choice:"4241:AD_SERVING" enumAccept:"scalar" enumStyle:"scalar" enumAlias:"4237:Ad Serving" enumCase:"fold"`
	}
	type body struct {
		Rows  []*fee          `json:"rows"`
		Named map[string]*fee `json:"named"`
	}
	for _, output := range []bool{false, true} {
		reflector := Reflector{Output: output, Annotate: func(_ string, f reflect.StructField) (string, any) {
			if f.Name == "Value" {
				return "Fee type", "AD_SERVING"
			}
			return "", nil
		}}
		doc, err := reflector.Compile(Request{Type: reflect.TypeFor[body]()})
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(doc)
		if err != nil {
			t.Fatal(err)
		}
		text := string(encoded)
		if !strings.Contains(text, "Fee type") || !strings.Contains(text, "AD_SERVING") || strings.Contains(text, `"Integer"`) || strings.Contains(text, `"String"`) {
			t.Fatalf("metadata/storage: %s", text)
		}
		if strings.Contains(text, "Ad Serving") == output {
			t.Fatalf("input alias/output canonical distinction: %s", text)
		}
		if !strings.Contains(text, `"type":"null"`) {
			t.Fatalf("nonpointer null: %s", text)
		}
		again, err := reflector.Compile(Request{Type: reflect.TypeFor[body]()})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(doc, again) {
			t.Fatal("schema generation retained mutable state")
		}
	}
}

func TestScalarUnionSelectedRepresentation(t *testing.T) {
	type row struct {
		Value choice.ScalarValue `json:"value"`
	}
	walker, err := shape.CompileRewrite(reflect.TypeFor[row]())
	if err != nil {
		t.Fatal(err)
	}
	metadata, err := walker.RewriteRepresentation(func(_ string, _ reflect.Type, f *shape.FieldMetadata) (*shape.FieldMetadata, error) {
		if f.Name == "Value" {
			return &shape.FieldMetadata{Name: "Fee", Tag: `json:"fee" choice:"4237:AD_SERVING" enumAccept:"scalar" enumStyle:"scalar" enumInteger:"passthrough" enumUnknownString:"preserve"`}, nil
		}
		return nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	fields, err := metadata.JSONFields()
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 1 {
		t.Fatal(fields)
	}
	field := fields[0]
	for _, output := range []bool{false, true} {
		doc, err := (Reflector{Output: output}).Compile(Request{Type: field.Field.ReflectedType, Field: &field, Representation: metadata.FieldShapeByIndex(field.Field.Index)})
		if err != nil {
			t.Fatal(err)
		}
		alternatives, ok := doc["anyOf"].([]any)
		if !ok || len(alternatives) != 3 {
			t.Fatalf("selected union: %v", doc)
		}
	}
}
