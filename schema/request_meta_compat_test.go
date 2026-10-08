package schema

import (
	"encoding/json"
	"testing"
)

func TestRequestMetaObjectAcceptsJune2025CompatibilityMetadata(t *testing.T) {
	var meta RequestMetaObject
	err := json.Unmarshal([]byte(`{"io.modelcontextprotocol/protocolVersion":"2025-06-18","progressToken":0}`), &meta)
	if err != nil {
		t.Fatalf("expected June 2025 metadata to decode: %v", err)
	}
}

func TestRequestMetaObjectRequiresCapabilitiesForJulyProtocol(t *testing.T) {
	var meta RequestMetaObject
	err := json.Unmarshal([]byte(`{"io.modelcontextprotocol/protocolVersion":"2026-07-28"}`), &meta)
	if err == nil {
		t.Fatal("expected July metadata without client capabilities to fail")
	}
}
