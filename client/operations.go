package client

import (
	"context"
	"github.com/viant/jsonrpc"
	"github.com/viant/jsonrpc/transport"
	"github.com/viant/mcp-protocol/schema"
)

type Operations interface {
	transport.Notifier
	transport.Sequencer
	ListRoots(ctx context.Context, request *jsonrpc.TypedRequest[*schema.ListRootsRequest]) (*schema.ListRootsResult, *jsonrpc.Error)
	CreateMessage(ctx context.Context, params *jsonrpc.TypedRequest[*schema.CreateMessageRequest]) (*schema.CreateMessageResult, *jsonrpc.Error)
	Elicit(ctx context.Context, params *jsonrpc.TypedRequest[*schema.ElicitRequest]) (*schema.ElicitResult, *jsonrpc.Error)
	Implements(method string) bool
	Init(ctx context.Context, capabilities *schema.ClientCapabilities)
}
