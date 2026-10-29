package schema

import (
	"bytes"
	"encoding/json"
)

// MarshalJSON preserves the output schema vocabulary on the wire.
func (s ToolOutputSchema) MarshalJSON() ([]byte, error) {
	type plain ToolOutputSchema
	return marshalToolSchema(plain(s), s.AdditionalProperties, "type", "properties", "required", "$schema")
}

// UnmarshalJSON preserves output vocabulary, including schemas with no type.
func (s *ToolOutputSchema) UnmarshalJSON(data []byte) error {
	type plain ToolOutputSchema
	var value plain
	if err := unmarshalToolSchemaJSON(data, &value); err != nil {
		return err
	}
	var fields map[string]interface{}
	if err := unmarshalToolSchemaJSON(data, &fields); err != nil {
		return err
	}
	for _, key := range []string{"type", "properties", "required", "$schema"} {
		delete(fields, key)
	}
	value.AdditionalProperties = fields
	*s = ToolOutputSchema(value)
	return nil
}

func marshalToolSchema(value interface{}, extensions interface{}, authoritative ...string) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var fields map[string]interface{}
	if err := unmarshalToolSchemaJSON(data, &fields); err != nil {
		return nil, err
	}
	if flag, ok := extensions.(bool); ok {
		fields["additionalProperties"] = flag
	} else if extensions != nil {
		data, err := json.Marshal(extensions)
		if err != nil {
			return nil, err
		}
		var extra map[string]interface{}
		if err := unmarshalToolSchemaJSON(data, &extra); err != nil {
			return nil, err
		}
		for _, key := range authoritative {
			delete(extra, key)
		}
		for key, value := range extra {
			fields[key] = value
		}
	}
	return json.Marshal(fields)
}

// Keep arbitrary JSON numbers exact through schema persistence.
func unmarshalToolSchemaJSON(raw []byte, target interface{}) error {
	if !json.Valid(raw) {
		var value interface{}
		return json.Unmarshal(raw, &value)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	return decoder.Decode(target)
}
