package server

import (
	"context"
	"github.com/viant/jsonrpc"
	"github.com/viant/mcp-protocol/schema"
)

// Operations represents implementation interface
type Operations interface {
	Initialize(ctx context.Context, init *schema.InitializeRequestParams, result *schema.InitializeResult)

	ListResources(ctx context.Context, request *schema.ListResourcesRequest) (*schema.ListResourcesResult, *jsonrpc.Error)

	ListResourceTemplates(ctx context.Context, request *schema.ListResourceTemplatesRequest) (*schema.ListResourceTemplatesResult, *jsonrpc.Error)

	ReadResource(ctx context.Context, request *schema.ReadResourceRequest) (*schema.ReadResourceResult, *jsonrpc.Error)

	Subscribe(ctx context.Context, request *schema.SubscribeRequest) (*schema.SubscribeResult, *jsonrpc.Error)

	Unsubscribe(ctx context.Context, request *schema.UnsubscribeRequest) (*schema.UnsubscribeResult, *jsonrpc.Error)

	ListTools(ctx context.Context, request *schema.ListToolsRequest) (*schema.ListToolsResult, *jsonrpc.Error)

	CallTool(ctx context.Context, request *schema.CallToolRequest) (*schema.CallToolResult, *jsonrpc.Error)

	ListPrompts(ctx context.Context, request *schema.ListPromptsRequest) (*schema.ListPromptsResult, *jsonrpc.Error)

	GetPrompt(ctx context.Context, request *schema.GetPromptRequest) (*schema.GetPromptResult, *jsonrpc.Error)

	Complete(ctx context.Context, request *schema.CompleteRequest) (*schema.CompleteResult, *jsonrpc.Error)
}
