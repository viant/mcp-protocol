package schema

const (
	LatestProtocolVersion   = "2025-06-18"
	TokenProgressContextKey = tokenProgress("TokenProgress")
	McpSessionContextKey    = mcpSessionId("MCPSessionId")
)

type tokenProgress string
type mcpSessionId string
