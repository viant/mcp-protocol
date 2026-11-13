package server

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/viant/mcp-protocol/client"
	"github.com/viant/mcp-protocol/schema"
)

type capabilityClient struct{ client.Operations }

func (*capabilityClient) Init(context.Context, *schema.ClientCapabilities) {}

func TestRegisteredToolsPreserveNotificationCapabilities(t *testing.T) {
	for _, tc := range []struct {
		name     string
		declared *bool
		want     string
	}{
		{name: "unspecified", want: `{}`},
		{name: "enabled", declared: capabilityBool(true), want: `{"listChanged":true}`},
		{name: "disabled", declared: capabilityBool(false), want: `{"listChanged":false}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := NewDefaultHandler(nil, nil, &capabilityClient{})
			if tc.declared != nil {
				h.ServerCapabilities = &schema.ServerCapabilities{Tools: &schema.ServerCapabilitiesTools{ListChanged: tc.declared}}
			}
			h.RegisterTool(&ToolEntry{Metadata: schema.Tool{Name: "report_menu"}})
			var initialized schema.InitializeResult
			h.Initialize(t.Context(), &schema.InitializeRequestParams{}, &initialized)
			var discovered schema.DiscoverResult
			h.Discover(t.Context(), &discovered)
			for _, caps := range []schema.ServerCapabilities{initialized.Capabilities, discovered.Capabilities} {
				require.NotNil(t, caps.Tools)
				raw, err := json.Marshal(caps.Tools)
				require.NoError(t, err)
				require.JSONEq(t, tc.want, string(raw))
			}
		})
	}
}

func capabilityBool(value bool) *bool { return &value }
