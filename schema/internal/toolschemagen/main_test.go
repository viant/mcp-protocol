package main

import (
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
