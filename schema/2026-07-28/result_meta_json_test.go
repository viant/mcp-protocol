package schema

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestResultMetadataExtensionsRoundTrip(t *testing.T) {
	raw := []byte(`{"ui":{"resourceUri":"ui://example"},"example/flag":true,"example/null":null,"io.modelcontextprotocol/serverInfo":{"name":"server","version":"1"}}`)
	var meta ResultMetaObject
	if err := json.Unmarshal(raw, &meta); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(meta)
	if err != nil {
		t.Fatal(err)
	}
	var want, got map[string]interface{}
	_ = json.Unmarshal(raw, &want)
	_ = json.Unmarshal(encoded, &got)
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("metadata changed: %s", encoded)
	}
	meta.AdditionalProperties.(map[string]interface{})["io.modelcontextprotocol/serverInfo"] = "not the typed identity"
	encoded, err = json.Marshal(meta)
	if err != nil {
		t.Fatal(err)
	}
	_ = json.Unmarshal(encoded, &got)
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("extension replaced typed identity: %s", encoded)
	}
	if err := json.Unmarshal([]byte("{}"), &meta); err != nil {
		t.Fatal(err)
	}
	if meta.AdditionalProperties != nil || meta.IoModelcontextprotocolServerInfo != nil {
		t.Fatal("reuse retained stale metadata")
	}
}

func TestResultMetadataInvalidCatchall(t *testing.T) {
	if _, err := json.Marshal(ResultMetaObject{AdditionalProperties: "invalid"}); err == nil {
		t.Fatal("invalid catchall accepted")
	}
	var meta ResultMetaObject
	if err := json.Unmarshal([]byte("[]"), &meta); err == nil {
		t.Fatal("array accepted as metadata")
	}
}
