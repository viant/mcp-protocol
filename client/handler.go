package client

import (
	"context"
	"github.com/viant/jsonrpc"
)

// Handler extends Operations with support for JSON-RPC notifications.
type Handler interface {
	Operations

	OnNotification(ctx context.Context, notification *jsonrpc.Notification)
}
