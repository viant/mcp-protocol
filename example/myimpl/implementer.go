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

// MyMCPServer is a sample MCP implementer embedding the default Base.
type MyMCPServer struct {
	*server.DefaultHandler
}

// ListResources implements the resources/list method.
func (i *MyMCPServer) ListResources(
	ctx context.Context,
	req *schema.ListResourcesRequest,
) (*schema.ListResourcesResult, *jsonrpc.Error) {
	// TODO: return actual resources
	return &schema.ListResourcesResult{}, nil
}

// Implements indicates which methods this implementer supports.
func (i *MyMCPServer) Implements(method string) bool {
	return method == schema.MethodResourcesList
}

// NewMCPServer returns a factory for MyMCPServer.
func NewMCPServer() server.NewHandler {
	return func(
		ctx context.Context,
		notifier transport.Notifier,
		log logger.Logger,
		client client.Operations,
	) (server.Handler, error) {
		base := server.NewDefaultHandler(notifier, log, client)
		return &MyMCPServer{DefaultHandler: base}, nil
	}
}
