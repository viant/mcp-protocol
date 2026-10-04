package schema

import (
	"encoding/json"
	"fmt"
)

// MarshalJSON flattens the JSON Schema vocabulary held in the generated
// catchall. Typed core fields always win; extension keys cannot replace them.
func (j ToolInputSchema) MarshalJSON() ([]byte, error) {
	type core struct {
		Schema     *string                   `json:"$schema,omitempty"`
		Properties ToolInputSchemaProperties `json:"properties,omitempty"`
		Required   []string                  `json:"required,omitempty"`
		Type       string                    `json:"type"`
	}
	encoded, err := json.Marshal(core{Schema: j.Schema, Properties: j.Properties, Required: j.Required, Type: j.Type})
	if err != nil {
		return nil, err
	}
	fields := map[string]interface{}{}
	if err = json.Unmarshal(encoded, &fields); err != nil {
		return nil, err
	}
	if j.AdditionalProperties != nil {
		extras, ok := j.AdditionalProperties.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("ToolInputSchema catchall must contain a JSON Schema vocabulary map")
		}
		for key, value := range extras {
			switch key {
			case "$schema", "properties", "required", "type", "AdditionalProperties":
				continue
			}
			fields[key] = value
		}
	}
	return json.Marshal(fields)
}
