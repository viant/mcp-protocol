package myimpl

import (
	"context"
	"github.com/viant/jsonrpc"
	"github.com/viant/jsonrpc/transport"
	"github.com/viant/mcp-protocol/client"
	"github.com/viant/mcp-protocol/logger"
	"github.com/viant/mcp-protocol/schema"
	"github.com/viant/mcp-protocol/server"
)

// MyImplementer is a sample MCP implementer embedding the default Base.
type MyImplementer struct {
	*server.DefaultImplementer
}

// ListResources implements the resources/list method.
func (i *MyImplementer) ListResources(
	ctx context.Context,
	req *schema.ListResourcesRequest,
) (*schema.ListResourcesResult, *jsonrpc.Error) {
	// TODO: return actual resources
	return &schema.ListResourcesResult{}, nil
}

// Implements indicates which methods this implementer supports.
func (i *MyImplementer) Implements(method string) bool {
	return method == schema.MethodResourcesList
}

// NewMyImplementer returns a factory for MyImplementer.
func NewMyImplementer() server.NewImplementer {
	return func(
		ctx context.Context,
		notifier transport.Notifier,
		log logger.Logger,
		client client.Operations,
	) (server.Implementer, error) {
		base := server.NewDefaultImplementer(notifier, log, client)
		return &MyImplementer{DefaultImplementer: base}, nil
	}
}
