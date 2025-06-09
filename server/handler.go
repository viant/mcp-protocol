package server

import (
	"context"
	"github.com/viant/jsonrpc"
	"github.com/viant/jsonrpc/transport"
	"github.com/viant/mcp-protocol/client"
	"github.com/viant/mcp-protocol/logger"
)

// Handler represents a protocol implementer.
type Handler interface {
	Operations

	OnNotification(ctx context.Context, notification *jsonrpc.Notification)

	// Implements checks if the method is implemented.
	Implements(method string) bool
}

// NewHandler creates new handler implementer.
type NewHandler func(ctx context.Context, notifier transport.Notifier, logger logger.Logger, client client.Operations) (Handler, error)
