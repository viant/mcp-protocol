package schema

const (
	LatestProtocolVersion   = "2025-11-25"
	TokenProgressContextKey = tokenProgress("TokenProgress")
	McpSessionContextKey    = mcpSessionId("MCPSessionId")
)

type tokenProgress string
type mcpSessionId string
