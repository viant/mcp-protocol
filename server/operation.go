package server

import (
	"context"
	"github.com/viant/jsonrpc"
	"github.com/viant/mcp-protocol/schema"
)

// Operations lists all JSON-RPC methods that an MCP handler may implement.
type Operations interface {
	Initialize(ctx context.Context, init *schema.InitializeRequestParams, result *schema.InitializeResult)

	ListResources(ctx context.Context, request *jsonrpc.TypedRequest[*schema.ListResourcesRequest]) (*schema.ListResourcesResult, *jsonrpc.Error)

	ListResourceTemplates(ctx context.Context, request *jsonrpc.TypedRequest[*schema.ListResourceTemplatesRequest]) (*schema.ListResourceTemplatesResult, *jsonrpc.Error)

	ReadResource(ctx context.Context, request *jsonrpc.TypedRequest[*schema.ReadResourceRequest]) (*schema.ReadResourceResult, *jsonrpc.Error)

	Subscribe(ctx context.Context, request *jsonrpc.TypedRequest[*schema.SubscribeRequest]) (*schema.SubscribeResult, *jsonrpc.Error)

	Unsubscribe(ctx context.Context, request *jsonrpc.TypedRequest[*schema.UnsubscribeRequest]) (*schema.UnsubscribeResult, *jsonrpc.Error)

	ListTools(ctx context.Context, request *jsonrpc.TypedRequest[*schema.ListToolsRequest]) (*schema.ListToolsResult, *jsonrpc.Error)

	CallTool(ctx context.Context, request *jsonrpc.TypedRequest[*schema.CallToolRequest]) (*schema.CallToolResult, *jsonrpc.Error)

	ListPrompts(ctx context.Context, request *jsonrpc.TypedRequest[*schema.ListPromptsRequest]) (*schema.ListPromptsResult, *jsonrpc.Error)

	GetPrompt(ctx context.Context, request *jsonrpc.TypedRequest[*schema.GetPromptRequest]) (*schema.GetPromptResult, *jsonrpc.Error)

	Complete(ctx context.Context, request *jsonrpc.TypedRequest[*schema.CompleteRequest]) (*schema.CompleteResult, *jsonrpc.Error)
}
