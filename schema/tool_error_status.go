package schema

import (
	"encoding/json"
	"math"
	"math/big"
)

// ToolErrorHTTPStatusMetaKey is an optional Viant extension, not an MCP standard.
// It preserves a producer's typed HTTP failure status outside its public body.
// It classifies a tool failure only; it never grants access or replaces isError.
const ToolErrorHTTPStatusMetaKey = "io.viant/http-status"

// A three-digit status needs only a short numeric representation. This bound
// applies to optional metadata lexemes, before any arbitrary-precision work.
const maxToolErrorHTTPStatusNumberBytes = 64

// SetToolErrorHTTPStatus attaches a typed 400–599 status to an error result.
// Metadata is copied so unrelated results sharing it remain unchanged. Explicit
// content, structuredContent, server identity and other extensions are retained.
func SetToolErrorHTTPStatus(result *CallToolResult, status int) bool {
	if result == nil || result.IsError == nil || !*result.IsError || status < 400 || status > 599 {
		return false
	}
	meta := ResultMetaObject{}
	fields := map[string]interface{}{}
	if result.Meta != nil {
		meta = *result.Meta
		if result.Meta.AdditionalProperties != nil {
			original, ok := result.Meta.AdditionalProperties.(map[string]interface{})
			if !ok {
				return false
			}
			for k, v := range original {
				fields[k] = v
			}
		}
	}
	fields[ToolErrorHTTPStatusMetaKey] = status
	meta.AdditionalProperties = fields
	result.Meta = &meta
	return true
}

// ToolErrorHTTPStatus accepts only numeric integral 400–599 metadata on an
// isError result. Strings, fractional values and metadata on successes are ignored.
func ToolErrorHTTPStatus(result *CallToolResult) (int, bool) {
	if result == nil || result.IsError == nil || !*result.IsError || result.Meta == nil {
		return 0, false
	}
	fields, ok := result.Meta.AdditionalProperties.(map[string]interface{})
	if !ok {
		return 0, false
	}
	var number float64
	switch value := fields[ToolErrorHTTPStatusMetaKey].(type) {
	case int:
		number = float64(value)
	case int8:
		number = float64(value)
	case int16:
		number = float64(value)
	case int32:
		number = float64(value)
	case int64:
		number = float64(value)
	case uint:
		number = float64(value)
	case uint8:
		number = float64(value)
	case uint16:
		number = float64(value)
	case uint32:
		number = float64(value)
	case uint64:
		number = float64(value)
	case float32:
		number = float64(value)
	case float64:
		number = value
	case json.Number:
		if len(value.String()) == 0 || len(value.String()) > maxToolErrorHTTPStatusNumberBytes {
			return 0, false
		}
		approximate, err := value.Float64()
		if err != nil || approximate < 400 || approximate > 599 {
			return 0, false
		}
		if !json.Valid([]byte(value.String())) {
			return 0, false
		}
		exact, ok := new(big.Rat).SetString(value.String())
		if !ok || !exact.IsInt() || !exact.Num().IsInt64() {
			return 0, false
		}
		status := exact.Num().Int64()
		if status < 400 || status > 599 {
			return 0, false
		}
		return int(status), true
	default:
		return 0, false
	}
	if number < 400 || number > 599 || math.Trunc(number) != number {
		return 0, false
	}
	return int(number), true
}
