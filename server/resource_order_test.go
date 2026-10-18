package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/viant/jsonrpc"
	"github.com/viant/mcp-protocol/schema"
)

func TestDefaultResourceCatalogIsStableForCursorPagination(t *testing.T) {
	h := NewDefaultHandler(nil, nil, nil)
	for _, uri := range []string{"window://studio/z", "window://studio/a", "window://studio/m"} {
		require.NoError(t, h.RegisterResource(schema.Resource{Uri: uri, Name: uri}, nil))
	}
	for _, uri := range []string{"window://studio/{z}", "window://studio/{a}"} {
		h.RegisterResourceTemplate(schema.ResourceTemplate{UriTemplate: uri, Name: uri}, nil)
	}
	for range 32 {
		listed, rpcErr := h.ListResources(context.Background(), nil)
		require.Nil(t, rpcErr)
		require.Equal(t, schema.ResultTypeComplete, listed.ResultType)
		require.Equal(t, schema.ListResourcesResultCacheScopePrivate, listed.CacheScope)
		require.Equal(t, []string{"window://studio/a", "window://studio/m", "window://studio/z"}, []string{listed.Resources[0].Uri, listed.Resources[1].Uri, listed.Resources[2].Uri})
		templates, rpcErr := h.ListResourceTemplates(context.Background(), nil)
		require.Nil(t, rpcErr)
		require.Equal(t, []string{"window://studio/{a}", "window://studio/{z}"}, []string{templates.ResourceTemplates[0].UriTemplate, templates.ResourceTemplates[1].UriTemplate})
	}
}

func TestMissingResourceReadUsesInvalidParams(t *testing.T) {
	h := NewDefaultHandler(nil, nil, nil)
	_, rpcErr := h.ReadResource(context.Background(), &jsonrpc.TypedRequest[*schema.ReadResourceRequest]{Request: &schema.ReadResourceRequest{Params: schema.ReadResourceRequestParams{Uri: "window://studio/missing"}}})
	require.NotNil(t, rpcErr)
	require.Equal(t, -32602, rpcErr.Code)
}
