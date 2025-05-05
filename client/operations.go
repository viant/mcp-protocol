package client

import (
	"context"
	"github.com/viant/jsonrpc"
	"github.com/viant/mcp-protocol/schema"
)

type Operations interface {
	ListRoots(ctx context.Context, params *schema.ListRootsRequestParams) (*schema.ListRootsResult, *jsonrpc.Error)
	CreateMessage(ctx context.Context, params *schema.CreateMessageRequestParams) (*schema.CreateMessageResult, *jsonrpc.Error)
}
