package schema

import "encoding/json"

// MarshalJSON preserves request metadata extensions while keeping standard
// protocol fields authoritative. It complements RequestMetaObject.UnmarshalJSON.
func (j RequestMetaObject) MarshalJSON() ([]byte, error) {
	type Plain RequestMetaObject
	raw, err := json.Marshal(Plain(j))
	if err != nil {
		return nil, err
	}
	var fields map[string]interface{}
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, err
	}
	var extensions map[string]interface{}
	if j.AdditionalProperties != nil {
		raw, err := json.Marshal(j.AdditionalProperties)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &extensions); err != nil {
			return nil, err
		}
	}
	for key, value := range extensions {
		switch key {
		case "io.modelcontextprotocol/clientCapabilities", "io.modelcontextprotocol/clientInfo", "io.modelcontextprotocol/logLevel", "io.modelcontextprotocol/protocolVersion", "progressToken":
			continue
		}
		fields[key] = value
	}
	return json.Marshal(fields)
}
