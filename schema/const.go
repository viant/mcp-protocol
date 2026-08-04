package schema

const (
	// LatestProtocolVersion is the preferred protocol when both peers support it.
	LatestProtocolVersion = "2026-07-28"
	// LegacyProtocolVersion is retained for explicit negotiated downgrade.
	LegacyProtocolVersion   = "2025-11-25"
	TokenProgressContextKey = tokenProgress("TokenProgress")
	McpSessionContextKey    = mcpSessionId("MCPSessionId")
)

// SupportedProtocolVersions is ordered by preference.
var SupportedProtocolVersions = []string{LatestProtocolVersion, LegacyProtocolVersion}

type tokenProgress string
type mcpSessionId string
