package schema

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestChoiceSchemaExactIntegerWire(t *testing.T) {
	for _, value := range []any{ToolInputSchema{Type: "object", Properties: ToolInputSchemaProperties{"mode": {"type": "integer", "enum": []any{int64(9007199254740993)}}}}, ToolOutputSchema{Type: "object", Properties: ToolInputSchemaProperties{"mode": {"type": "integer", "enum": []any{int64(9007199254740993)}}}}} {
		raw, e := json.Marshal(value)
		if e != nil {
			t.Fatal(e)
		}
		if !strings.Contains(string(raw), "9007199254740993") {
			t.Fatalf("typed enum rounded on wire: %s", raw)
		}
	}
}

func TestChoiceSchemaRoundTripExactIntegerVocabulary(t *testing.T) {
	raw := []byte(`{"type":"object","properties":{"mode":{"type":"integer","enum":[9007199254740993,9007199254740994]}},"allOf":[{"properties":{"other":{"enum":[9007199254740993]}}}]}`)
	var input ToolInputSchema
	var output ToolOutputSchema
	for _, target := range []any{&input, &output} {
		if e := json.Unmarshal(raw, target); e != nil {
			t.Fatal(e)
		}
		encoded, e := json.Marshal(target)
		if e != nil {
			t.Fatal(e)
		}
		if strings.Count(string(encoded), "9007199254740993") != 2 || !strings.Contains(string(encoded), "9007199254740994") {
			t.Fatalf("enum vocabulary rounded: %s", encoded)
		}
	}
}

func TestChoiceSchemaAdapterRejectsTrailingJSON(t *testing.T) {
	raw := []byte(`{"type":"object"}{"type":"integer"}`)
	var input ToolInputSchema
	var output ToolOutputSchema
	if input.UnmarshalJSON(raw) == nil || output.UnmarshalJSON(raw) == nil {
		t.Fatal("schema metadata accepted trailing JSON")
	}
}
