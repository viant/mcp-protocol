package server

import (
	"context"
	"github.com/viant/jsonrpc"
	"github.com/viant/jsonrpc/transport"
	"github.com/viant/mcp-protocol/client"
	"github.com/viant/mcp-protocol/logger"
)

// Server represents a implementer implementer
type Server interface {
	Operations

	OnNotification(ctx context.Context, notification *jsonrpc.Notification)

	// Implements checks if the method is implemented
	Implements(method string) bool
}

// NewServer creates new implementer
type NewServer func(ctx context.Context, notifier transport.Notifier, logger logger.Logger, client client.Operations) (Server, error)
