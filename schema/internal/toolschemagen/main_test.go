package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRewriteToolSchemaCarriers(t *testing.T) {
	source := []byte("package schema\ntype ToolInputSchema struct { AdditionalProperties interface{} `mapstructure:\",remain\"` }; type ToolOutputSchema struct { AdditionalProperties interface{} `mapstructure:\",remain\"` }; type Other struct { AdditionalProperties interface{} }\n")
	got, err := rewrite(source)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(got), `json:"-" yaml:"-" mapstructure:",remain"`) != 2 {
		t.Fatalf("missing tool tags: %s", got)
	}
	again, err := rewrite(got)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(again) {
		t.Fatal("not idempotent")
	}
}
func TestRewriteRejectsGeneratorShapeChange(t *testing.T) {
	if _, err := rewrite([]byte("package schema")); err == nil {
		t.Fatal("missing carriers accepted")
	}
}

func TestSchemaDecodeUsesExactNumericAdapter(t *testing.T) {
	source := []byte("package schema\nimport \"encoding/json\"\ntype ToolInputSchema struct { AdditionalProperties interface{} };type ToolOutputSchema struct { AdditionalProperties interface{} };func (j *ToolInputSchema) UnmarshalJSON(value []byte)error { var plain interface{};return json.Unmarshal(value,&plain) }\n")
	result, err := rewrite(source)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(result, []byte("unmarshalToolSchemaJSON(value, &plain)")) {
		t.Fatalf("root decoder did not retain exact schema vocabulary: %s", result)
	}
	again, err := rewrite(result)
	if err != nil || !bytes.Equal(result, again) {
		t.Fatalf("schema metadata generation not stable: %v", err)
	}
}

func TestToolCallDecoderRemainsStandardJSON(t *testing.T) {
	for _, owner := range []string{"CallToolRequestParams", "CallToolRequest", "CallToolResult", "CallToolResultResponse", "CallToolResultResponseResult"} {
		source := []byte("package schema\nimport \"encoding/json\"\ntype ToolInputSchema struct{AdditionalProperties interface{}};type ToolOutputSchema struct{AdditionalProperties interface{}};type " + owner + " struct{};func(j *" + owner + ")UnmarshalJSON(raw []byte)error{var value interface{};return json.Unmarshal(raw,&value)}")
		result, err := rewrite(source)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(result, []byte("json.Unmarshal(raw, &value)")) {
			t.Fatalf("tool call decoding changed: %s", result)
		}
	}
}
