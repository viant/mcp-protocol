package authorization

import (
	"context"
	"github.com/viant/mcp-protocol/oauth2/meta"
	"golang.org/x/oauth2"
)

type ProtectedResourceTokenSource interface {
	ProtectedResourceToken(ctx context.Context, protectedResource *meta.ProtectedResourceMetadata, scope string) (*oauth2.Token, error)
}

type IdTokenSource interface {
	IdToken(ctx context.Context, token *oauth2.Token, protectedResource *meta.ProtectedResourceMetadata) (*oauth2.Token, error)
}
