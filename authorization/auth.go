package authorization

import (
	"github.com/viant/mcp-protocol/oauth2/meta"
)

// Token carries authentication credentials.
type Token struct {
	Token string `json:"token"`
}

// WithMeta extracts authorization metadata from JSON-RPC params.
type WithMeta struct {
	Name     string `json:"name"`
	AuthMeta struct {
		Authorization *Token `json:"authorization,omitempty"`
	} `json:"_meta,omitempty"`
}

// Authorization defines per-resource aI do uthorization requirements.
type Authorization struct {
	ProtectedResourceMetadata *meta.ProtectedResourceMetadata `json:"protectedResourceMetadata"`
	RequiredScopes            []string                        `json:"requiredScopes"`
	UseIdToken                bool                            `json:"useIdToken,omitempty"`
}
