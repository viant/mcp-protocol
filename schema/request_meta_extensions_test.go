package schema

import (
	"encoding/json"
	"testing"
)

func TestRequestMetaExtensionsRoundTripAndCannotOverrideProtocolFields(t *testing.T) {
	raw := []byte(`{"io.modelcontextprotocol/protocolVersion":"2025-06-18","viant.datly/component":{"id":"example","revision":"release-1"},"AdditionalProperties":{"opaque":true},"IoModelcontextprotocolProtocolVersion":"opaque"}`)
	var meta RequestMetaObject
	if err := json.Unmarshal(raw, &meta); err != nil {
		t.Fatal(err)
	}
	extensions := meta.AdditionalProperties.(map[string]interface{})
	if extensions["viant.datly/component"] == nil {
		t.Fatal("extension lost")
	}
	extensions["io.modelcontextprotocol/protocolVersion"] = "forged"
	extensions["progressToken"] = "forged"
	encoded, err := json.Marshal(meta)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["viant.datly/component"] == nil || decoded["io.modelcontextprotocol/protocolVersion"] != "2025-06-18" || decoded["progressToken"] != nil {
		t.Fatalf("metadata authority changed or extension lost: %s", encoded)
	}
	if decoded["AdditionalProperties"] == nil || decoded["IoModelcontextprotocolProtocolVersion"] != "opaque" {
		t.Fatalf("unknown metadata was reinterpreted as a Go field: %s", encoded)
	}
}
