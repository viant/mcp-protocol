package schema

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	legacy "github.com/viant/mcp-protocol/schema/2025-11-25"
)

func TestProtocolVersionPreference(t *testing.T) {
	require.Len(t, SupportedProtocolVersions, 2)
	assert.Equal(t, "2026-07-28", LatestProtocolVersion)
	assert.Equal(t, LatestProtocolVersion, SupportedProtocolVersions[0])
	assert.Equal(t, LegacyProtocolVersion, SupportedProtocolVersions[1])
}

func TestLegacyInitializeCompatibility(t *testing.T) {
	request := InitializeRequestParams{
		Capabilities:    ClientCapabilities{},
		ClientInfo:      *NewImplementation("legacy-client", "1.0.0"),
		ProtocolVersion: LegacyProtocolVersion,
	}

	data, err := json.Marshal(request)
	require.NoError(t, err)

	var decoded legacy.InitializeRequestParams
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, LegacyProtocolVersion, decoded.ProtocolVersion)
	assert.Equal(t, "legacy-client", decoded.ClientInfo.Name)
}

func TestJulyRequestCompatibilityFields(t *testing.T) {
	var request ListPromptsRequest
	require.NoError(t, json.Unmarshal([]byte(`{
        "id": 1,
        "jsonrpc": "2.0",
        "method": "prompts/list",
        "params": {
          "_meta": {
            "io.modelcontextprotocol/clientCapabilities": {},
            "io.modelcontextprotocol/protocolVersion": "2026-07-28"
          },
          "cursor": "next"
        }
    }`), &request))

	require.NotNil(t, request.Params.Cursor)
	assert.Equal(t, "next", *request.Params.Cursor)
	require.NotNil(t, request.PaginatedRequestParams)
	assert.Equal(t, request.Params.Cursor, request.PaginatedRequestParams.Cursor)
}

func TestJulyInputRequiredDecodesThroughRootResults(t *testing.T) {
	state := "opaque-state"
	for name, target := range map[string]interface{}{
		"tool":     &CallToolResult{},
		"prompt":   &GetPromptResult{},
		"resource": &ReadResourceResult{},
	} {
		t.Run(name, func(t *testing.T) {
			raw := []byte(`{"resultType":"input_required","inputRequests":{"approval":{"type":"boolean"}},"requestState":"opaque-state"}`)
			require.NoError(t, json.Unmarshal(raw, target))
			switch value := target.(type) {
			case *CallToolResult:
				assert.Equal(t, state, *value.RequestState)
			case *GetPromptResult:
				assert.Equal(t, state, *value.RequestState)
			case *ReadResourceResult:
				assert.Equal(t, state, *value.RequestState)
			}
		})
	}
}
