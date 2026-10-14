package namespaces

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrCursor = errors.New("invalid or stale namespace catalog cursor")

// Paginate follows the existing skills catalog's snapshot/offset cursor
// convention. The owner MUST authorize, filter and deterministically sort the
// collection before calling this helper. It neither grants access nor filters
// resources, and clients must treat returned cursors as opaque.
func Paginate[T any](entries []T, cursor *string, pageSize int) ([]T, *string, error) {
	if pageSize <= 0 {
		return nil, nil, fmt.Errorf("catalog page size must be positive")
	}
	raw, err := json.Marshal(entries)
	if err != nil {
		return nil, nil, err
	}
	revision := fmt.Sprintf("%x", sha256.Sum256(raw))
	start := 0
	if cursor != nil {
		decoded, err := base64.RawURLEncoding.DecodeString(*cursor)
		parts := strings.Split(string(decoded), ":")
		if err != nil || len(parts) != 2 || parts[0] != revision {
			return nil, nil, ErrCursor
		}
		start, err = strconv.Atoi(parts[1])
		if err != nil || start <= 0 || start >= len(entries) || start%pageSize != 0 || strconv.Itoa(start) != parts[1] {
			return nil, nil, ErrCursor
		}
	}
	end := start + min(pageSize, len(entries)-start)
	page := append([]T{}, entries[start:end]...)
	if end == len(entries) {
		return page, nil, nil
	}
	next := base64.RawURLEncoding.EncodeToString([]byte(revision + ":" + strconv.Itoa(end)))
	return page, &next, nil
}
