package schema

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestEmbeddedBlobPreservesEmptyAndBinary(t *testing.T) {
	for _, value := range [][]byte{nil, {0, 255}, []byte(`{"arbitrary":"JSON"}`)} {
		content := NewEmbeddedBlob("datly://response/body", "application/octet-stream", value)
		data, err := json.Marshal(content)
		if err != nil {
			t.Fatal(err)
		}
		var raw map[string]any
		if err = json.Unmarshal(data, &raw); err != nil {
			t.Fatal(err)
		}
		resource := raw["resource"].(map[string]any)
		if _, ok := resource["text"]; ok {
			t.Fatal("blob also emitted text")
		}
		decoded, err := base64.StdEncoding.DecodeString(resource["blob"].(string))
		if err != nil || !bytes.Equal(value, decoded) {
			t.Fatal(string(data))
		}
		var roundtrip EmbeddedResource
		if err = json.Unmarshal(data, &roundtrip); err != nil {
			t.Fatal(err)
		}
		again, err := json.Marshal(roundtrip)
		if err != nil || !bytes.Equal(data, again) {
			t.Fatal(string(again), err)
		}
	}
}
