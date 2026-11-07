package jsonschema

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/viant/tagly/format"
	"github.com/xeipuuv/gojsonschema"
)

func TestAliasPrioritySchemaAllowsDiscardedCanonical(t *testing.T) {
	property := map[string]any{"type": "integer"}
	object := map[string]any{"type": "object", "properties": map[string]any{"flightId": property}}
	required := projectInputAliases(object, "flightId", property, &format.Tag{InputNames: "CampaignFlightId", InputNamesPriority: "alias", InputNamesCase: "fold"}, true)
	require.False(t, required)
	result, err := gojsonschema.Validate(gojsonschema.NewGoLoader(object), gojsonschema.NewGoLoader(map[string]any{"flightId": "discarded", "CampaignFlightId": 2}))
	require.NoError(t, err)
	require.True(t, result.Valid(), "%v", result.Errors())
	require.Equal(t, "integer", property["type"], "canonical source shape must not be mutated")
}
