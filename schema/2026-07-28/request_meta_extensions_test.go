package schema

import (
	"encoding/json"
	"testing"
)

func TestRequestMetadataExtensionsJulyDialect(t *testing.T) {
	var meta RequestMetaObject
	err := json.Unmarshal([]byte(`{"io.modelcontextprotocol/protocolVersion":"2026-07-28","io.modelcontextprotocol/clientCapabilities":{},"example/unknown":{"enabled":false,"value":null}}`), &meta)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(meta)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	if string(fields["example/unknown"]) != `{"enabled":false,"value":null}` {
		t.Fatalf("opaque metadata lost: %s", encoded)
	}
}
