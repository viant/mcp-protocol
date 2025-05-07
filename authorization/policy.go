package authorization

// Policy holds OAuth2/OIDC configuration for fine-grained control.

type Policy struct {
	// Global resource protection metadata (mutually exclusive with Tools/Resources)
	Global *Authorization `json:"global,omitempty"`
	// ExcludeURI skips middleware on matching paths
	ExcludeURI string `json:"excludeURI,omitempty"`
	// Per-tool authorization metadata
	Tools map[string]*Authorization `json:"tools,omitempty"`
	// Per-tenant authorization metadata (reserved for future use)
	Resources map[string]*Authorization `json:"resources,omitempty"`
}

// IsFineGrained reports whether this config uses fine-grained (tool/resource) control.
func (a *Policy) IsFineGrained() bool {
	if a == nil {
		return false
	}
	// If Global is set, use spec-based global protection
	if a.Global != nil {
		return false
	}
	return len(a.Tools) > 0 || len(a.Resources) > 0
}
