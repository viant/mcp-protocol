package schema

import (
	"encoding/json"
	"fmt"
)

// MarshalJSON keeps extension keys at the _meta object's top level. The typed
// server identity is authoritative and cannot be overridden by the catchall.
func (m ResultMetaObject) MarshalJSON() ([]byte, error) {
	fields := map[string]interface{}{}
	if m.AdditionalProperties != nil {
		extras, ok := m.AdditionalProperties.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("ResultMetaObject extensions must be a map")
		}
		for key, value := range extras {
			if key != "io.modelcontextprotocol/serverInfo" {
				fields[key] = value
			}
		}
	}
	if m.IoModelcontextprotocolServerInfo != nil {
		fields["io.modelcontextprotocol/serverInfo"] = m.IoModelcontextprotocolServerInfo
	}
	return json.Marshal(fields)
}

// UnmarshalJSON retains unknown metadata instead of silently dropping it.
func (m *ResultMetaObject) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	next := ResultMetaObject{}
	if raw, ok := fields["io.modelcontextprotocol/serverInfo"]; ok {
		if err := json.Unmarshal(raw, &next.IoModelcontextprotocolServerInfo); err != nil {
			return err
		}
		delete(fields, "io.modelcontextprotocol/serverInfo")
	}
	if len(fields) > 0 {
		extras := map[string]interface{}{}
		for key, raw := range fields {
			var value interface{}
			if err := json.Unmarshal(raw, &value); err != nil {
				return err
			}
			extras[key] = value
		}
		next.AdditionalProperties = extras
	}
	*m = next
	return nil
}
