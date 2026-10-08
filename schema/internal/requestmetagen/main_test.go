package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestGeneratorPreservesOriginalSchemaAndOtherDefinitions(t *testing.T) {
	input := []byte(`{"$defs":{"RequestMetaObject":{"type":"object","properties":{"standard":{"type":"string"}}},"Unrelated":{"const":"unchanged"}}}`)
	original := bytes.Clone(input)
	output, err := generatorSchema(input)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(input, original) {
		t.Fatal("generator modified the supplied schema")
	}
	var result struct {
		Defs map[string]map[string]json.RawMessage `json:"$defs"`
	}
	if err = json.Unmarshal(output, &result); err != nil {
		t.Fatal(err)
	}
	if string(result.Defs["RequestMetaObject"]["additionalProperties"]) != "{}" {
		t.Fatal("missing generator hint")
	}
	if string(result.Defs["Unrelated"]["const"]) != `"unchanged"` {
		t.Fatal("unrelated schema changed")
	}
}
func TestGeneratorRejectsNewProtocolRestrictions(t *testing.T) {
	for _, constraint := range []string{`false`, `{"type":"string"}`} {
		input := []byte(`{"$defs":{"RequestMetaObject":{"type":"object","additionalProperties":` + constraint + `}}}`)
		if _, err := generatorSchema(input); err == nil {
			t.Fatal("silently relaxed protocol restriction")
		}
	}
}
