package schema

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSkillRequiredFieldsAndUnion(t *testing.T) {
	for _, raw := range []string{`{}`, `{"uri":"skill://demo/SKILL.md","frontmatter":{}}`, `{"uri":"skill://demo/SKILL.md","frontmatter":{},"resources":null}`, `{"uri":"skill://demo/SKILL.md","frontmatter":{},"resources":[{"uri":"x","digest":"y"}]}`} {
		var skill Skill
		require.Error(t, json.Unmarshal([]byte(raw), &skill), raw)
	}
	var skill Skill
	require.NoError(t, json.Unmarshal([]byte(`{"uri":"skill://demo/SKILL.md","frontmatter":{"name":"demo","description":"demo","future":9007199254740993},"resources":"dynamic"}`), &skill))
	require.True(t, skill.Resources.Dynamic)
	raw, err := json.Marshal(skill)
	require.NoError(t, err)
	require.Contains(t, string(raw), "9007199254740993")
}
