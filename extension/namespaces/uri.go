package namespaces

import (
	"errors"
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"
)

var ErrURI = errors.New("invalid namespace URI")

// URI encodes an opaque namespace identifier without imposing an application
// ID format (for example hexadecimal, UUID, DNS name, or case normalization).
func URI(id string) (string, error) {
	if id == "" || !utf8.ValidString(id) || strings.IndexFunc(id, unicode.IsControl) >= 0 {
		return "", ErrURI
	}
	const hex = "0123456789ABCDEF"
	var result strings.Builder
	result.WriteString("ns://")
	for i := 0; i < len(id); i++ {
		c := id[i]
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '.' || c == '_' || c == '~' {
			result.WriteByte(c)
		} else {
			result.WriteByte('%')
			result.WriteByte(hex[c>>4])
			result.WriteByte(hex[c&15])
		}
	}
	return result.String(), nil
}

// ParseURI returns the exact opaque ID. Collection paths and resource URIs are
// distinct contracts and are not accepted as namespace references.
func ParseURI(value string) (string, error) {
	if !strings.HasPrefix(value, "ns://") {
		return "", ErrURI
	}
	id, err := url.PathUnescape(strings.TrimPrefix(value, "ns://"))
	if err != nil {
		return "", ErrURI
	}
	canonical, err := URI(id)
	if err != nil || canonical != value {
		return "", ErrURI
	}
	return id, nil
}
