package extension

// Package extension contains optional protocol helpers that are not part of
// the formal MCP spec but can be shared by tools and clients for common
// behaviors (e.g., pagination/truncation hints).

// Continuation describes a generic pagination/truncation hint that tools can
// attach to their responses. Presence of this object signals that additional
// data may be retrieved by following the provided range or page token.
type Continuation struct {
	// HasMore reports that more data is available beyond what was returned.
	HasMore bool `json:"hasMore,omitempty"`
	// Remaining is the estimated number of bytes (or units) not returned.
	Remaining int `json:"remaining,omitempty"`
	// Returned is the number of bytes (or units) included in this response.
	Returned int `json:"returned,omitempty"`
	// NextRange provides the next suggested byte/line window for continuation.
	NextRange *RangeHint `json:"nextRange,omitempty"`
	// PageToken is an opaque continuation token alternative to NextRange.
	PageToken string `json:"pageToken,omitempty"`
	// Mode optionally records the preview mode applied (e.g., head, tail).
	Mode string `json:"mode,omitempty"`
	// Binary indicates the returned content is a placeholder/preview for binary data.
	Binary bool `json:"binary,omitempty"`
}

// RangeHint defines the next slice to fetch, expressed as bytes or lines.
type RangeHint struct {
	Bytes *ByteRange `json:"bytes,omitempty"`
	Lines *LineRange `json:"lines,omitempty"`
}

// ByteRange specifies a half-open [Offset, Offset+Length) window.
type ByteRange struct {
	Offset int `json:"offset,omitempty"`
	Length int `json:"length,omitempty"`
}

// LineRange specifies a 1-based start line and count window.
type LineRange struct {
	Start int `json:"start,omitempty"`
	Count int `json:"count,omitempty"`
}
