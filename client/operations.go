package client

import (
	"context"
	"github.com/viant/jsonrpc"
	"github.com/viant/jsonrpc/transport"
	"github.com/viant/mcp-protocol/schema"
)

type Operations interface {
	transport.Notifier
	ListRoots(ctx context.Context, request *jsonrpc.TypedRequest[*schema.ListRootsRequest]) (*schema.ListRootsResult, *jsonrpc.Error)
	CreateMessage(ctx context.Context, params *jsonrpc.TypedRequest[*schema.CreateMessageRequest]) (*schema.CreateMessageResult, *jsonrpc.Error)
	Elicit(ctx context.Context, params *jsonrpc.TypedRequest[*schema.ElicitRequest]) (*schema.ElicitResult, *jsonrpc.Error)
	CreateUserInteraction(ctx context.Context, params *jsonrpc.TypedRequest[*schema.CreateUserInteractionRequest]) (*schema.CreateUserInteractionResult, *jsonrpc.Error)
	Implements(method string) bool
	Init(ctx context.Context, capabilities *schema.ClientCapabilities)
}
