package schema

// JSON-RPC method names for MCP protocol
const (
	MethodInitialize                  = "initialize"
	MethodPing                        = "notifications/ping"
	MethodPong                        = "notifications/pong"
	MethodResourcesList               = "resources/list"
	MethodResourcesTemplatesList      = "resources/templates/list"
	MethodResourcesRead               = "resources/read"
	MethodSubscribe                   = "resources/subscribe"
	MethodUnsubscribe                 = "resources/unsubscribe"
	MethodPromptsList                 = "prompts/list"
	MethodPromptsGet                  = "prompts/get"
	MethodToolsList                   = "tools/list"
	MethodToolsCall                   = "tools/call"
	MethodComplete                    = "complete"
	MethodLoggingSetLevel             = "logging/setLevel"
	MethodNotificationInitialized     = "notifications/initialized"
	MethodNotificationResourceUpdated = "notifications/resources/updated"
	MethodNotificationMessage         = "notifications/message"
	MethodNotificationCancel          = "cancel"
	MethodRootsList                   = "roots/list"
	MethodSamplingCreateMessage       = "sampling/createMessage"
)
