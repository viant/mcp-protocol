package schema

import "github.com/viant/mcp-protocol/jsonvalue"

// Keep normalization confined to tool-call generic payloads. Schema vocabulary,
// typed IDs and declared numeric fields retain their own decoding contracts.
func normalizeToolCallNumbers(target interface{}) {
	switch value := target.(type) {
	case *CallToolRequestParams:
		jsonvalue.Normalize(value.Arguments)
		jsonvalue.Normalize(value.InputResponses)
	case *CallToolResult:
		value.StructuredContent = jsonvalue.Normalize(value.StructuredContent)
		jsonvalue.Normalize(value.InputRequests)
		for index, item := range value.Content {
			value.Content[index] = jsonvalue.Normalize(item)
		}
	case *CallToolResultResponseResult:
		value.StructuredContent = jsonvalue.Normalize(value.StructuredContent)
		jsonvalue.Normalize(value.InputRequests)
		for index, item := range value.Content {
			value.Content[index] = jsonvalue.Normalize(item)
		}
	}
}
