package schema

// Standard Streamable HTTP headers introduced by the 2026-07-28 protocol.
const (
	HeaderProtocolVersion = "Mcp-Protocol-Version"
	HeaderMethod          = "Mcp-Method"
	HeaderName            = "Mcp-Name"
	HeaderParamPrefix     = "Mcp-Param-"
)

// Standard protocol error codes used while validating HTTP requests.
const (
	ErrorCodeHeaderMismatch             = -32020
	ErrorCodeUnsupportedProtocolVersion = -32022
)
