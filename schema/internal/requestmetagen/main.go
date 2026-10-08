// Command requestmetagen regenerates the extensible request metadata carrier
// without replacing other maintained protocol compatibility projections.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if err := generate(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func generate() error {
	const source = "schema-2026-07-28.json"
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	data, err = generatorSchema(data)
	if err != nil {
		return err
	}
	dir, err := os.MkdirTemp("", "mcp-request-meta-generate-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	input := filepath.Join(dir, source)
	if err := os.WriteFile(input, data, 0600); err != nil {
		return err
	}
	generated := filepath.Join(dir, "types.go")
	cmd := exec.Command("go", "run", "github.com/atombender/go-jsonschema@v0.20.0", input, "-p", "schema", "-o", generated)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	output, err := os.ReadFile(generated)
	if err != nil {
		return err
	}
	block, err := carrier(output)
	if err != nil {
		return err
	}
	block = bytes.ReplaceAll(block, []byte("AdditionalProperties interface{} `mapstructure:\",remain\"`"), []byte("AdditionalProperties interface{} `json:\"-\" yaml:\"-\" mapstructure:\",remain\"`"))
	// Go field names are not wire keys. Preserve all unknown metadata as data,
	// including keys coinciding with generated Go names or the carrier field.
	block = bytes.ReplaceAll(block, []byte(`delete(raw, st.Field(i).Name)
		delete(raw, strings.Split(st.Field(i).Tag.Get("json"), ",")[0])`), []byte(`key := strings.Split(st.Field(i).Tag.Get("json"), ",")[0]
		if key != "" && key != "-" {
			delete(raw, key)
		}`))
	// The explicit July package retains its strict July requirements; the root
	// carrier additionally supports older negotiated dialects.
	if err := replaceCarrier(filepath.Join("2026-07-28", "types.go"), block); err != nil {
		return err
	}
	// Preserve negotiated June compatibility while July requires capabilities.
	old := `if _, ok := raw["io.modelcontextprotocol/clientCapabilities"]; raw != nil && !ok {
		return fmt.Errorf("field io.modelcontextprotocol/clientCapabilities in RequestMetaObject: required")
	}
	if _, ok := raw["io.modelcontextprotocol/protocolVersion"]; raw != nil && !ok {
		return fmt.Errorf("field io.modelcontextprotocol/protocolVersion in RequestMetaObject: required")
	}`
	replacement := `protocolVersion, _ := raw["io.modelcontextprotocol/protocolVersion"].(string)
	if raw != nil && protocolVersion == "" {
		return fmt.Errorf("field io.modelcontextprotocol/protocolVersion in RequestMetaObject: required")
	}
	// Per-request capabilities are a July protocol requirement. Older MCP
	// clients, including the 2025-06-18 compatibility dialect, send only the
	// negotiated protocol version (and optionally a progress token).
	if protocolVersion == LatestProtocolVersion {
		if _, ok := raw["io.modelcontextprotocol/clientCapabilities"]; !ok {
			return fmt.Errorf("field io.modelcontextprotocol/clientCapabilities in RequestMetaObject: required")
		}
	}`
	if !strings.Contains(string(block), old) {
		return fmt.Errorf("generated metadata requirements changed; review compatibility projection")
	}
	block = bytes.Replace(block, []byte(old), []byte(replacement), 1)
	return replaceCarrier("types.go", block)
}

func replaceCarrier(path string, block []byte) error {
	current, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	prior, err := carrier(current)
	if err != nil {
		return err
	}
	current = bytes.Replace(current, prior, block, 1)
	return os.WriteFile(path, current, 0644)
}

func carrier(data []byte) ([]byte, error) {
	start := bytes.Index(data, []byte("type RequestMetaObject struct {"))
	if start < 0 {
		return nil, fmt.Errorf("request metadata carrier not generated")
	}
	end := bytes.Index(data[start:], []byte("// Common params for any request."))
	if end < 0 {
		return nil, fmt.Errorf("request metadata carrier boundary missing")
	}
	return data[start : start+end], nil
}

// generatorSchema makes JSON Schema's implicit open-object semantics explicit
// solely for code generation. The vendored protocol schema remains unchanged.
func generatorSchema(data []byte) ([]byte, error) {
	var document map[string]json.RawMessage
	if err := json.Unmarshal(data, &document); err != nil {
		return nil, err
	}
	var definitions map[string]json.RawMessage
	if err := json.Unmarshal(document["$defs"], &definitions); err != nil {
		return nil, err
	}
	var metadata map[string]json.RawMessage
	if err := json.Unmarshal(definitions["RequestMetaObject"], &metadata); err != nil {
		return nil, err
	}
	if metadata == nil {
		return nil, fmt.Errorf("request metadata schema is missing")
	}
	if declared, ok := metadata["additionalProperties"]; ok {
		value := bytes.TrimSpace(declared)
		if !bytes.Equal(value, []byte("{}")) && !bytes.Equal(value, []byte("true")) {
			return nil, fmt.Errorf("request metadata extension constraints changed; review generator")
		}
	}
	metadata["additionalProperties"] = json.RawMessage(`{}`)
	var err error
	if definitions["RequestMetaObject"], err = json.Marshal(metadata); err != nil {
		return nil, err
	}
	if document["$defs"], err = json.Marshal(definitions); err != nil {
		return nil, err
	}
	return json.MarshalIndent(document, "", "  ")
}
