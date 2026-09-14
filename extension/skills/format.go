// Package skills validates Agent Skills metadata and builds SEP-2640 inventories.
// It neither fetches remote resources nor activates or authorizes skill content.
package skills

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"
	"unicode"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

// Frontmatter preserves every JSON-compatible author field without coercing
// scalars into strings. Non-JSON YAML values cannot be advertised verbatim.
func Frontmatter(data []byte) (map[string]interface{}, error) {
	if !utf8.Valid(data) {
		return nil, fmt.Errorf("SKILL.md must be UTF-8")
	}
	lines := bytes.Split(data, []byte("\n"))
	if len(lines) < 3 || string(bytes.TrimSuffix(lines[0], []byte("\r"))) != "---" {
		return nil, fmt.Errorf("SKILL.md must begin with YAML frontmatter")
	}
	end := 1
	for ; end < len(lines); end++ {
		if string(bytes.TrimSuffix(lines[end], []byte("\r"))) == "---" {
			break
		}
	}
	if end == len(lines) {
		return nil, fmt.Errorf("unterminated skill frontmatter")
	}
	var result map[string]interface{}
	if err := yaml.Unmarshal(bytes.Join(lines[1:end], []byte("\n")), &result); err != nil {
		return nil, err
	}
	// Round-trip only JSON-compatible values so the registry and wire agree.
	raw, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("non-JSON skill frontmatter: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err = decoder.Decode(&result); err != nil {
		return nil, err
	}
	if err = ValidateFrontmatter(result); err != nil {
		return nil, err
	}
	return result, nil
}

func ValidateFrontmatter(front map[string]interface{}) error {
	name, ok := front["name"].(string)
	if !ok || utf8.RuneCountInString(name) < 1 || utf8.RuneCountInString(name) > 64 || strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") || strings.Contains(name, "--") {
		return fmt.Errorf("invalid skill name")
	}
	for _, r := range name {
		if r != '-' && !unicode.IsLower(r) && !unicode.IsDigit(r) {
			return fmt.Errorf("invalid skill name")
		}
	}
	description, ok := front["description"].(string)
	if !ok || strings.TrimSpace(description) == "" || utf8.RuneCountInString(description) > 1024 {
		return fmt.Errorf("invalid skill description")
	}
	for _, key := range []string{"license", "allowed-tools", "compatibility"} {
		if value, exists := front[key]; exists {
			text, ok := value.(string)
			if !ok {
				return fmt.Errorf("skill %s must be a string", key)
			}
			if key == "compatibility" && (text == "" || utf8.RuneCountInString(text) > 500) {
				return fmt.Errorf("invalid skill compatibility")
			}
		}
	}
	if value, exists := front["metadata"]; exists {
		metadata, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("skill metadata must be a string map")
		}
		for _, v := range metadata {
			if _, ok := v.(string); !ok {
				return fmt.Errorf("skill metadata values must be strings")
			}
		}
	}
	return nil
}

// Root validates structure independent of URI scheme; a scheme never declares
// a skill. The caller must explicitly register this entry on its own server.
func Root(uri, name string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil || u.Scheme == "" || u.Host == "" || u.User != nil || u.Opaque != "" || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || strings.ContainsAny(u.Path, "\\%") || u.RawPath != "" || path.Clean(u.Path) != u.Path || !strings.HasSuffix(u.Path, "/SKILL.md") {
		return "", fmt.Errorf("invalid skill URI %q", uri)
	}
	root := strings.TrimSuffix(uri, "/SKILL.md")
	skillPath := strings.TrimSuffix(u.Path, "/SKILL.md")
	last := u.Host
	if skillPath != "" {
		last = path.Base(skillPath)
	}
	if last != name {
		return "", fmt.Errorf("skill URI root must end in frontmatter name %q", name)
	}
	return root, nil
}
