package schema

// JSON-RPC method names for MCP protocol
const (
	MethodInitialize                  = "initialize"
	MethodPing                        = "ping"
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
	MethodNotificationInitialized     = "notification/initialized"
	MethodNotificationMessage         = "notification/message"
	MethodNotificationCancel          = "notification/cancel"
	MethodNotificationResourceUpdated = "notifications/resources/updated"
	MethodRootsList                   = "roots/list"
	MethodSamplingCreateMessage       = "sampling/createMessage"
)
