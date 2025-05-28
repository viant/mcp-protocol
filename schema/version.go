package schema

import "time"

const dateLayout = "2006-01-02"

// IsProtocolNewer returns true if `current` represents a later protocol date
// than `reference`. The date strings must be in ISO layout "YYYY-MM-DD". If
// either value cannot be parsed the function returns false.
//
// Example:
//
//	IsProtocolNewer("2025-06-21", "2025-03-26") == true
func IsProtocolNewer(current, reference string) bool {
	refT, err := time.Parse(dateLayout, reference)
	if err != nil {
		return false
	}
	curT, err := time.Parse(dateLayout, current)
	if err != nil {
		return false
	}
	return curT.After(refT)
}
