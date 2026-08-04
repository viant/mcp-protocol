package schema

import legacy "github.com/viant/mcp-protocol/schema/2025-11-25"

// Deprecated: these aliases preserve the pre-July root-package API. New code
// should import schema/2025-11-25 explicitly when speaking the legacy protocol.
type (
	CallToolRequestParamsMeta                  = legacy.CallToolRequestParamsMeta
	CancelTaskRequest                          = legacy.CancelTaskRequest
	CancelTaskRequestParams                    = legacy.CancelTaskRequestParams
	CancelTaskResult                           = legacy.CancelTaskResult
	ClientCapabilitiesRoots                    = legacy.ClientCapabilitiesRoots
	ClientCapabilitiesTasks                    = legacy.ClientCapabilitiesTasks
	ClientCapabilitiesTasksRequests            = legacy.ClientCapabilitiesTasksRequests
	ClientCapabilitiesTasksRequestsElicitation = legacy.ClientCapabilitiesTasksRequestsElicitation
	ClientCapabilitiesTasksRequestsSampling    = legacy.ClientCapabilitiesTasksRequestsSampling
	ClientResult                               = legacy.ClientResult
	CompleteRequestParamsMeta                  = legacy.CompleteRequestParamsMeta
	CreateMessageRequestParamsMeta             = legacy.CreateMessageRequestParamsMeta
	CreateTaskResult                           = legacy.CreateTaskResult
	ElicitRequestFormParamsMeta                = legacy.ElicitRequestFormParamsMeta
	ElicitRequestURLParamsMeta                 = legacy.ElicitRequestURLParamsMeta
	ElicitationCompleteNotification            = legacy.ElicitationCompleteNotification
	ElicitationCompleteNotificationParams      = legacy.ElicitationCompleteNotificationParams
	GetPromptRequestParamsMeta                 = legacy.GetPromptRequestParamsMeta
	GetTaskPayloadRequest                      = legacy.GetTaskPayloadRequest
	GetTaskPayloadRequestParams                = legacy.GetTaskPayloadRequestParams
	GetTaskPayloadResult                       = legacy.GetTaskPayloadResult
	GetTaskRequest                             = legacy.GetTaskRequest
	GetTaskRequestParams                       = legacy.GetTaskRequestParams
	GetTaskResult                              = legacy.GetTaskResult
	InitializeRequestParamsMeta                = legacy.InitializeRequestParamsMeta
	InitializedNotification                    = legacy.InitializedNotification
	ListTasksRequest                           = legacy.ListTasksRequest
	ListTasksResult                            = legacy.ListTasksResult
	PaginatedRequestParamsMeta                 = legacy.PaginatedRequestParamsMeta
	PingRequest                                = legacy.PingRequest
	ReadResourceRequestParamsMeta              = legacy.ReadResourceRequestParamsMeta
	RelatedTaskMetadata                        = legacy.RelatedTaskMetadata
	RequestParamsMeta                          = legacy.RequestParamsMeta
	ResourceRequestParamsMeta                  = legacy.ResourceRequestParamsMeta
	RootsListChangedNotification               = legacy.RootsListChangedNotification
	ServerCapabilitiesTasks                    = legacy.ServerCapabilitiesTasks
	ServerCapabilitiesTasksRequests            = legacy.ServerCapabilitiesTasksRequests
	ServerCapabilitiesTasksRequestsTools       = legacy.ServerCapabilitiesTasksRequestsTools
	ServerRequest                              = legacy.ServerRequest
	SetLevelRequestParamsMeta                  = legacy.SetLevelRequestParamsMeta
	SubscribeRequest                           = legacy.SubscribeRequest
	SubscribeRequestParams                     = legacy.SubscribeRequestParams
	SubscribeRequestParamsMeta                 = legacy.SubscribeRequestParamsMeta
	Task                                       = legacy.Task
	TaskAugmentedRequestParams                 = legacy.TaskAugmentedRequestParams
	TaskAugmentedRequestParamsMeta             = legacy.TaskAugmentedRequestParamsMeta
	TaskMetadata                               = legacy.TaskMetadata
	TaskStatus                                 = legacy.TaskStatus
	TaskStatusNotification                     = legacy.TaskStatusNotification
	TaskStatusNotificationParams               = legacy.TaskStatusNotificationParams
	ToolExecution                              = legacy.ToolExecution
	ToolExecutionTaskSupport                   = legacy.ToolExecutionTaskSupport
	URLElicitationRequiredError                = legacy.URLElicitationRequiredError
	URLElicitationRequiredErrorError           = legacy.URLElicitationRequiredErrorError
	URLElicitationRequiredErrorErrorData       = legacy.URLElicitationRequiredErrorErrorData
	UnsubscribeRequest                         = legacy.UnsubscribeRequest
	UnsubscribeRequestParams                   = legacy.UnsubscribeRequestParams
	UnsubscribeRequestParamsMeta               = legacy.UnsubscribeRequestParamsMeta
)

// ElicitRequestParams is the legacy superset retained by the root API. It can
// represent both July form and URL modes; the versioned July package exposes
// the strict union generated from the schema.
type ElicitRequestParams struct {
	Meta            *URLElicitRequestParamsMeta        `json:"_meta,omitempty" yaml:"_meta,omitempty" mapstructure:"_meta,omitempty"`
	ElicitationId   string                             `json:"elicitationId,omitempty" yaml:"elicitationId,omitempty" mapstructure:"elicitationId,omitempty"`
	Message         string                             `json:"message" yaml:"message" mapstructure:"message"`
	Mode            ElicitRequestParamsMode            `json:"mode,omitempty" yaml:"mode,omitempty" mapstructure:"mode,omitempty"`
	RequestedSchema ElicitRequestParamsRequestedSchema `json:"requestedSchema,omitempty" yaml:"requestedSchema,omitempty" mapstructure:"requestedSchema,omitempty"`
	Url             string                             `json:"url,omitempty" yaml:"url,omitempty" mapstructure:"url,omitempty"`
}

// SetLevelRequest remains available for legacy sessions while using the root
// LoggingLevel type expected by existing callers.
type SetLevelRequest struct {
	Id      RequestId             `json:"id" yaml:"id" mapstructure:"id"`
	Jsonrpc string                `json:"jsonrpc" yaml:"jsonrpc" mapstructure:"jsonrpc"`
	Method  string                `json:"method" yaml:"method" mapstructure:"method"`
	Params  SetLevelRequestParams `json:"params" yaml:"params" mapstructure:"params"`
}

type SetLevelRequestParams struct {
	Meta  *SetLevelRequestParamsMeta `json:"_meta,omitempty" yaml:"_meta,omitempty" mapstructure:"_meta,omitempty"`
	Level LoggingLevel               `json:"level" yaml:"level" mapstructure:"level"`
}

// InitializeRequest and related models remain available for negotiated legacy
// sessions. Their capability fields use the July-preferred root models so
// existing root-package callers remain source compatible.
type InitializeRequest struct {
	Id      RequestId               `json:"id" yaml:"id" mapstructure:"id"`
	Jsonrpc string                  `json:"jsonrpc" yaml:"jsonrpc" mapstructure:"jsonrpc"`
	Method  string                  `json:"method" yaml:"method" mapstructure:"method"`
	Params  InitializeRequestParams `json:"params" yaml:"params" mapstructure:"params"`
}

type InitializeRequestParams struct {
	Meta            *InitializeRequestParamsMeta `json:"_meta,omitempty" yaml:"_meta,omitempty" mapstructure:"_meta,omitempty"`
	Capabilities    ClientCapabilities           `json:"capabilities" yaml:"capabilities" mapstructure:"capabilities"`
	ClientInfo      Implementation               `json:"clientInfo" yaml:"clientInfo" mapstructure:"clientInfo"`
	ProtocolVersion string                       `json:"protocolVersion" yaml:"protocolVersion" mapstructure:"protocolVersion"`
}

type InitializeResult struct {
	Meta            map[string]interface{} `json:"_meta,omitempty" yaml:"_meta,omitempty" mapstructure:"_meta,omitempty"`
	Capabilities    ServerCapabilities     `json:"capabilities" yaml:"capabilities" mapstructure:"capabilities"`
	Instructions    *string                `json:"instructions,omitempty" yaml:"instructions,omitempty" mapstructure:"instructions,omitempty"`
	ProtocolVersion string                 `json:"protocolVersion" yaml:"protocolVersion" mapstructure:"protocolVersion"`
	ServerInfo      Implementation         `json:"serverInfo" yaml:"serverInfo" mapstructure:"serverInfo"`
}

// NewImplementation preserves the constructor exposed by the previous root
// schema while returning the July Implementation type.
func NewImplementation(name, version string) *Implementation {
	return &Implementation{Name: name, Version: version}
}

type (
	ElicitRequestElicitRequestParams                   interface{}
	ElicitRequestParamsMode                            string
	ElicitRequestParamsRequestedSchema                 = ElicitRequestFormParamsRequestedSchema
	ListPromptsRequestParams                           = PaginatedRequestParams
	ListResourceTemplatesRequestParams                 = PaginatedRequestParams
	ListResourcesRequestParams                         = PaginatedRequestParams
	ListToolsRequestParams                             = PaginatedRequestParams
	PingRequestParams                                  = RequestParams
	TaskStatusNotificationTaskStatusNotificationParams interface{}
	URLElicitRequestParamsMeta                         = ElicitRequestURLParamsMeta
)

// Deprecated: legacy enum values are retained for source compatibility.
const (
	ElicitRequestParamsModeForm ElicitRequestParamsMode = "form"
	ElicitRequestParamsModeUrl  ElicitRequestParamsMode = "url"

	TaskStatusWorking       = legacy.TaskStatusWorking
	TaskStatusInputRequired = legacy.TaskStatusInputRequired
	TaskStatusCompleted     = legacy.TaskStatusCompleted
	TaskStatusFailed        = legacy.TaskStatusFailed
	TaskStatusCancelled     = legacy.TaskStatusCancelled

	ToolExecutionTaskSupportForbidden = legacy.ToolExecutionTaskSupportForbidden
	ToolExecutionTaskSupportOptional  = legacy.ToolExecutionTaskSupportOptional
	ToolExecutionTaskSupportRequired  = legacy.ToolExecutionTaskSupportRequired
)
