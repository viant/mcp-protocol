package schema_test

import (
	"bytes"
	"encoding/json"
	root "github.com/viant/mcp-protocol/schema"
	june "github.com/viant/mcp-protocol/schema/2025-06-18"
	november "github.com/viant/mcp-protocol/schema/2025-11-25"
	july "github.com/viant/mcp-protocol/schema/2026-07-28"
	draft "github.com/viant/mcp-protocol/schema/draft"
	"testing"
)

func TestEmbeddedContentUnionAcrossProtocolVersions(t *testing.T) {
	versions := []struct {
		name    string
		blob    func([]byte) any
		text    func(string) any
		literal any
		both    any
		decode  func([]byte) (any, error)
	}{
		{name: "root", blob: func(data []byte) any {
			return root.NewEmbeddedBlob("datly://response/body", "application/octet-stream", data)
		}, text: func(text string) any { return root.NewEmbeddedText("datly://response/body", "text/plain", text) }, literal: root.EmbeddedResource{Type: "resource", Resource: root.EmbeddedResourceResource{Uri: "datly://response/body", Blob: ""}}, both: root.EmbeddedResource{Type: "resource", Resource: root.EmbeddedResourceResource{Uri: "datly://response/body", Blob: "YQ==", Text: "a"}}, decode: func(data []byte) (any, error) {
			var result root.EmbeddedResource
			err := json.Unmarshal(data, &result)
			return result, err
		}},
		{name: "june", blob: func(data []byte) any {
			return june.NewEmbeddedBlob("datly://response/body", "application/octet-stream", data)
		}, text: func(text string) any { return june.NewEmbeddedText("datly://response/body", "text/plain", text) }, literal: june.EmbeddedResource{Type: "resource", Resource: june.EmbeddedResourceResource{Uri: "datly://response/body", Blob: ""}}, both: june.EmbeddedResource{Type: "resource", Resource: june.EmbeddedResourceResource{Uri: "datly://response/body", Blob: "YQ==", Text: "a"}}, decode: func(data []byte) (any, error) {
			var result june.EmbeddedResource
			err := json.Unmarshal(data, &result)
			return result, err
		}},
		{name: "november", blob: func(data []byte) any {
			return november.NewEmbeddedBlob("datly://response/body", "application/octet-stream", data)
		}, text: func(text string) any { return november.NewEmbeddedText("datly://response/body", "text/plain", text) }, literal: november.EmbeddedResource{Type: "resource", Resource: november.EmbeddedResourceResource{Uri: "datly://response/body", Blob: ""}}, both: november.EmbeddedResource{Type: "resource", Resource: november.EmbeddedResourceResource{Uri: "datly://response/body", Blob: "YQ==", Text: "a"}}, decode: func(data []byte) (any, error) {
			var result november.EmbeddedResource
			err := json.Unmarshal(data, &result)
			return result, err
		}},
		{name: "july", blob: func(data []byte) any {
			return july.NewEmbeddedBlob("datly://response/body", "application/octet-stream", data)
		}, text: func(text string) any { return july.NewEmbeddedText("datly://response/body", "text/plain", text) }, literal: july.EmbeddedResource{Type: "resource", Resource: july.EmbeddedResourceResource{Uri: "datly://response/body", Blob: ""}}, both: july.EmbeddedResource{Type: "resource", Resource: july.EmbeddedResourceResource{Uri: "datly://response/body", Blob: "YQ==", Text: "a"}}, decode: func(data []byte) (any, error) {
			var result july.EmbeddedResource
			err := json.Unmarshal(data, &result)
			return result, err
		}},
		{name: "draft", blob: func(data []byte) any {
			return draft.NewEmbeddedBlob("datly://response/body", "application/octet-stream", data)
		}, text: func(text string) any { return draft.NewEmbeddedText("datly://response/body", "text/plain", text) }, literal: draft.EmbeddedResource{Type: "resource", Resource: draft.EmbeddedResourceResource{Uri: "datly://response/body", Blob: ""}}, both: draft.EmbeddedResource{Type: "resource", Resource: draft.EmbeddedResourceResource{Uri: "datly://response/body", Blob: "YQ==", Text: "a"}}, decode: func(data []byte) (any, error) {
			var result draft.EmbeddedResource
			err := json.Unmarshal(data, &result)
			return result, err
		}},
	}
	for _, version := range versions {
		t.Run(version.name, func(t *testing.T) {
			for _, test := range []struct {
				name  string
				value any
				kind  string
			}{{"empty literal", version.literal, "blob"}, {"empty blob", version.blob(nil), "blob"}, {"binary blob", version.blob([]byte{0, 255}), "blob"}, {"empty text", version.text(""), "text"}, {"text", version.text("hello"), "text"}} {
				t.Run(test.name, func(t *testing.T) {
					data, err := json.Marshal(test.value)
					if err != nil {
						t.Fatal(err)
					}
					var object map[string]any
					if err = json.Unmarshal(data, &object); err != nil {
						t.Fatal(err)
					}
					resource := object["resource"].(map[string]any)
					if _, ok := resource[test.kind]; !ok {
						t.Fatal(string(data))
					}
					other := "blob"
					if test.kind == "blob" {
						other = "text"
					}
					if _, ok := resource[other]; ok {
						t.Fatal("ambiguous alternative", string(data))
					}
					// Every destination version must preserve the exact alternative, even empty.
					for _, destination := range versions {
						decoded, err := destination.decode(data)
						if err != nil {
							t.Fatal(destination.name, err)
						}
						again, err := json.Marshal(decoded)
						if err != nil || !bytes.Equal(data, again) {
							t.Fatalf("%s roundtrip: %s / %s / %v", destination.name, data, again, err)
						}
					}
				})
			}
			if _, err := json.Marshal(version.both); err == nil {
				t.Fatal("ambiguous literal accepted")
			}
			for _, fields := range []string{`"blob":"","text":""`, `"blob":"YQ==","text":"a"`, `"blob":null`, `"text":null`, `"blob":7`, `"text":false`, `"mimeType":"application/octet-stream"`} {
				data := []byte(`{"type":"resource","resource":{"uri":"datly://response/body",` + fields + `}}`)
				if _, err := version.decode(data); err == nil {
					t.Fatal("invalid union accepted", string(data))
				}
			}
		})
	}
}
