package schema

import (
	"encoding/base64"
	"encoding/json"
	"github.com/viant/mcp-protocol/schema/internal/resourcecontent"
)

// NewEmbeddedBlob creates the blob alternative without interpreting payload
// bytes, including an empty blob. It is available in every protocol version.
func NewEmbeddedBlob(uri, media string, data []byte) EmbeddedResource {
	return EmbeddedResource{Type: "resource", Resource: EmbeddedResourceResource{Uri: uri, MimeType: &media, Blob: base64.StdEncoding.EncodeToString(data), contentSelection: resourcecontent.Blob}}
}

// NewEmbeddedText records explicit text intent, including empty text.
func NewEmbeddedText(uri, media, text string) EmbeddedResource {
	return EmbeddedResource{Type: "resource", Resource: EmbeddedResourceResource{Uri: uri, MimeType: &media, Text: text, contentSelection: resourcecontent.Text}}
}
func (r EmbeddedResourceResource) MarshalJSON() ([]byte, error) {
	selected, err := r.contentSelection.Resolve(r.Blob, r.Text)
	if err != nil {
		return nil, err
	}
	if selected == resourcecontent.Blob {
		return json.Marshal(BlobResourceContents{Meta: r.Meta, Uri: r.Uri, MimeType: r.MimeType, Blob: r.Blob})
	}
	return json.Marshal(TextResourceContents{Meta: r.Meta, Uri: r.Uri, MimeType: r.MimeType, Text: r.Text})
}
