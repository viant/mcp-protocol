package schema

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsProtocolNewer(t *testing.T) {
	newerTests := []struct {
		name      string
		current   string
		reference string
		expected  bool
	}{
		{name: "current newer", current: "2025-07-01", reference: "2025-03-26", expected: true},
		{name: "same date", current: "2025-03-26", reference: "2025-03-26", expected: false},
		{name: "current older", current: "2024-12-31", reference: "2025-03-26", expected: false},
		{name: "bad current", current: "bad", reference: "2025-03-26", expected: false},
		{name: "bad reference", current: "2025-06-21", reference: "bad", expected: false},
	}

	for _, tc := range newerTests {
		actual := IsProtocolNewer(tc.current, tc.reference)
		assert.EqualValues(t, tc.expected, actual, tc.name)
	}
}
