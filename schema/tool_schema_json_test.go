package schema_test

import (
	"encoding/json"
	"github.com/viant/mcp-protocol/schema"
	"reflect"
	"testing"
)

func TestToolSchemaVocabularyRoundTrip(t *testing.T) {
	for _, additional := range []string{"false", "true", `{"type":"string"}`} {
		t.Run(additional, func(t *testing.T) {
			input := `{"type":"object","properties":{"value":{"type":"string"}},"additionalProperties":` + additional + `,"$defs":{"choice":{"enum":["a","b"]}},"unevaluatedProperties":false}`
			inventory := `{"tools":[{"name":"example","inputSchema":` + input + `,"outputSchema":` + input + `}],"resultType":"complete","cacheScope":"private","ttlMs":0}`
			var result schema.ListToolsResult
			if err := json.Unmarshal([]byte(inventory), &result); err != nil {
				t.Fatal(err)
			}
			for _, value := range []interface{}{result.Tools[0].InputSchema, result.Tools[0].OutputSchema} {
				data, err := json.Marshal(value)
				if err != nil {
					t.Fatal(err)
				}
				var got, want interface{}
				if err := json.Unmarshal(data, &got); err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal([]byte(input), &want); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("schema vocabulary changed: got %s, want %s", data, input)
				}
			}
		})
	}
}

func TestToolOutputSchemaWithoutTypeRoundTrip(t *testing.T) {
	input := `{"$ref":"#/$defs/result","$defs":{"result":{"type":"object"}},"additionalProperties":false}`
	var value schema.ToolOutputSchema
	if err := json.Unmarshal([]byte(input), &value); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var got, want interface{}
	_ = json.Unmarshal(data, &got)
	_ = json.Unmarshal([]byte(input), &want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("output vocabulary changed: %s", data)
	}
}

func TestToolSchemaExtensionsCannotReplaceTypedFields(t *testing.T) {
	value := schema.ToolInputSchema{Type: "object", AdditionalProperties: map[string]interface{}{"type": "string", "additionalProperties": false}}
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]interface{}
	_ = json.Unmarshal(data, &got)
	if got["type"] != "object" || got["additionalProperties"] != false {
		t.Fatalf("invalid vocabulary: %s", data)
	}
}
