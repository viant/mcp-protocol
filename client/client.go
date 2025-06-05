package client

import (
	"context"
	"github.com/viant/jsonrpc"
)

type Client interface {
	Operations

	OnNotification(ctx context.Context, notification *jsonrpc.Notification)
}
