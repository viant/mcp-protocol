package schema

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestToolInputSchemaClosedVocabularyRoundTrip(t *testing.T) {
	source := []byte(`{"type":"object","additionalProperties":false,"properties":{"artifactId":{"type":"string"},"nested":{"type":"object","additionalProperties":false,"properties":{"value":{"type":"integer"}}}},"oneOf":[{"required":["artifactId"]},{"required":["recordingId"]}]}`)
	var value ToolInputSchema
	if e := json.Unmarshal(source, &value); e != nil {
		t.Fatal(e)
	}
	encoded, e := json.Marshal(value)
	if e != nil {
		t.Fatal(e)
	}
	var before, after map[string]interface{}
	_ = json.Unmarshal(source, &before)
	if e = json.Unmarshal(encoded, &after); e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("closed schema changed: %s", encoded)
	}
	var roundtrip ToolInputSchema
	if e = json.Unmarshal(encoded, &roundtrip); e != nil {
		t.Fatal(e)
	}
	if extras, ok := roundtrip.AdditionalProperties.(map[string]interface{}); !ok || extras["additionalProperties"] != false {
		t.Fatal("closed semantics lost")
	}
}
func TestToolInputSchemaTypedCoreWinsOverCatchall(t *testing.T) {
	vocabulary := "https://json-schema.org/draft/2020-12/schema"
	value := ToolInputSchema{Type: "object", Schema: &vocabulary, Required: []string{"artifactId"}, Properties: ToolInputSchemaProperties{"artifactId": {"type": "string"}}, AdditionalProperties: map[string]interface{}{"type": "array", "$schema": "untrusted", "required": []string{"path"}, "properties": map[string]interface{}{"path": true}, "additionalProperties": false, "oneOf": []interface{}{map[string]interface{}{"required": []string{"artifactId"}}}}}
	encoded, e := json.Marshal(value)
	if e != nil {
		t.Fatal(e)
	}
	var fields map[string]interface{}
	if e = json.Unmarshal(encoded, &fields); e != nil {
		t.Fatal(e)
	}
	if fields["type"] != "object" || fields["$schema"] != vocabulary || fields["additionalProperties"] != false || fields["AdditionalProperties"] != nil {
		t.Fatalf("typed fields replaced or catchall nested: %s", encoded)
	}
	required := fields["required"].([]interface{})
	if len(required) != 1 || required[0] != "artifactId" {
		t.Fatal("typed required overwritten")
	}
	properties := fields["properties"].(map[string]interface{})
	if _, ok := properties["artifactId"]; !ok {
		t.Fatal("typed properties overwritten")
	}
}
func TestToolInputSchemaReservedAbsentCoreCannotBeInjected(t *testing.T) {
	value := ToolInputSchema{Type: "object", AdditionalProperties: map[string]interface{}{"$schema": "injected", "required": []string{"path"}, "properties": map[string]interface{}{"path": true}}}
	encoded, e := json.Marshal(value)
	if e != nil {
		t.Fatal(e)
	}
	var fields map[string]interface{}
	_ = json.Unmarshal(encoded, &fields)
	if len(fields) != 1 || fields["type"] != "object" {
		t.Fatalf("absent typed core injected: %s", encoded)
	}
}
