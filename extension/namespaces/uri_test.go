package namespaces

import (
	"errors"
	"testing"
)

func TestNamespaceURIOpaqueRoundTrip(t *testing.T) {
	for _, id := range []string{"platform", "MixedCase_ID-1", "opaque/id?value#fragment", "space id", "团队", "literal%2F"} {
		t.Run(id, func(t *testing.T) {
			uri, err := URI(id)
			if err != nil {
				t.Fatal(err)
			}
			actual, err := ParseURI(uri)
			if err != nil || actual != id {
				t.Fatalf("roundtrip %q: %q, %v", uri, actual, err)
			}
		})
	}
	uri, _ := URI("MixedCase_ID-1")
	if uri != "ns://MixedCase_ID-1" {
		t.Fatalf("identifier changed: %q", uri)
	}
}

func TestNamespaceURIRejectsOtherContractsAndMalformedInput(t *testing.T) {
	for _, value := range []string{"", "ns://", "NS://platform", "window://platform/a", "/namespaces/platform/window", "ns://platform/window", "ns://platform?kind=window", "ns://platform#x", "ns://%", "ns://%ff", "ns://%00", "ns://%70latform", "ns://a%2fb"} {
		if _, err := ParseURI(value); !errors.Is(err, ErrURI) {
			t.Errorf("ParseURI(%q): %v", value, err)
		}
	}
	for _, id := range []string{"", "\x00", "\n", string([]byte{0xff})} {
		if _, err := URI(id); !errors.Is(err, ErrURI) {
			t.Errorf("URI(%q): %v", id, err)
		}
	}
}
