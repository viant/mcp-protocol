// Package resourcecontent owns blob/text alternative selection for every native
// MCP schema version. Wire fields remain in each version's generated models.
package resourcecontent

import "fmt"

type Selection uint8

const (
	Unspecified Selection = iota
	Blob
	Text
)

// FromObject preserves explicit empty alternatives during JSON round trips.
func FromObject(object map[string]any) (Selection, error) {
	_, blob := object["blob"]
	_, text := object["text"]
	if blob == text {
		return Unspecified, fmt.Errorf("embedded resource requires exactly one of blob or text")
	}
	if blob {
		if _, ok := object["blob"].(string); !ok {
			return Unspecified, fmt.Errorf("embedded blob must be a string")
		}
		return Blob, nil
	}
	if _, ok := object["text"].(string); !ok {
		return Unspecified, fmt.Errorf("embedded text must be a string")
	}
	return Text, nil
}

// Resolve rejects conflicting mutable fields. With no explicit alternative,
// nonempty text selects text and the all-empty value selects an empty blob.
// NewEmbeddedText or JSON decoding records explicit empty-text intent.
func (s Selection) Resolve(blob, text string) (Selection, error) {
	if blob != "" && text != "" || s == Blob && text != "" || s == Text && blob != "" {
		return Unspecified, fmt.Errorf("embedded resource cannot contain both text and blob")
	}
	if s != Unspecified {
		return s, nil
	}
	if text != "" {
		return Text, nil
	}
	return Blob, nil
}
