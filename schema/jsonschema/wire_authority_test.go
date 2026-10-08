package jsonschema

import (
	"reflect"
	"testing"
)

type recursiveCustomWireSource struct {
	Value string                     `json:"value"`
	Next  *recursiveCustomWireSource `json:"next,omitempty"`
}

func (*recursiveCustomWireSource) UnmarshalJSON([]byte) error {
	panic("schema compilation must not execute codecs")
}

type recursiveAuthoredWire recursiveCustomWireSource

func TestAuthoredWireAuthorityPreservesRecursiveCanonicalContract(t *testing.T) {
	source := reflect.TypeFor[recursiveCustomWireSource]()
	if _, err := (Reflector{}).Compile(Request{Type: source}); err == nil {
		t.Fatal("custom wire shape inferred without authority")
	}
	schema, err := (Reflector{WireTypes: map[reflect.Type]reflect.Type{source: reflect.TypeFor[recursiveAuthoredWire]()}}).Compile(Request{Type: source})
	if err != nil {
		t.Fatal(err)
	}
	properties := schema["properties"].(map[string]any)
	if properties["value"] == nil || properties["next"] == nil || schema["$defs"] == nil {
		t.Fatal("recursive authored contract incomplete", schema)
	}
}
func TestAuthoredWireAuthorityRejectsCyclesAndOrdinaryTypeOverrides(t *testing.T) {
	custom := reflect.TypeFor[recursiveCustomWireSource]()
	for _, authority := range []map[reflect.Type]reflect.Type{{custom: custom}, {reflect.TypeFor[string](): reflect.TypeFor[int]()}, {custom: nil}} {
		if _, err := (Reflector{WireTypes: authority}).Compile(Request{Type: custom}); err == nil {
			t.Fatal("invalid authority accepted")
		}
	}
}
