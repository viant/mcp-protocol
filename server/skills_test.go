package server

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/viant/jsonrpc"
	skillformat "github.com/viant/mcp-protocol/extension/skills"
	"github.com/viant/mcp-protocol/schema"
	"testing"
	"testing/fstest"
)

func TestSkillsRegistrySnapshotsAndStaleCursor(t *testing.T) {
	ctx := context.Background()
	d := NewDefaultHandler(nil, nil, nil)
	add := func(index int) {
		root := fmt.Sprintf("skill://team%d/demo", index)
		files := fstest.MapFS{"SKILL.md": {Data: []byte("---\nname: demo\ndescription: demo\n---\n")}}
		entry, err := (skillformat.Compiler{Source: files}).Compile(ctx, root+"/SKILL.md")
		require.NoError(t, err)
		require.Error(t, d.RegisterSkill(entry.Metadata()), "unproven manifest accepted")
		require.NoError(t, d.RegisterStaticSkill(entry))
		require.Error(t, d.RegisterStaticSkill(entry))
		entry.Metadata().Frontmatter["description"] = "mutated caller"
	}
	for i := 0; i < 33; i++ {
		add(i)
	}
	page, rpc := d.ListSkills(ctx, nil)
	require.Nil(t, rpc)
	require.Len(t, page.Skills, 32)
	require.Equal(t, "demo", page.Skills[0].Frontmatter["description"])
	add(33)
	_, rpc = d.ListSkills(ctx, &jsonrpc.TypedRequest[*schema.ListSkillsRequest]{Request: &schema.ListSkillsRequest{Params: schema.ListSkillsRequestParams{Cursor: page.NextCursor}}})
	require.NotNil(t, rpc)
	require.EqualValues(t, -32602, rpc.Code)
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	_, rpc = d.ListSkills(ctx, nil)
	require.NotNil(t, rpc)
}

func TestSkillsDynamicEntrypointRequiresReader(t *testing.T) {
	registry := NewRegistry()
	entry := schema.Skill{Uri: "skill://demo/SKILL.md", Frontmatter: map[string]interface{}{"name": "demo", "description": "demo"}, Resources: schema.SkillResources{Dynamic: true}}
	registry.RegisterResource(schema.Resource{Name: "demo", Uri: entry.Uri}, nil)
	require.Error(t, registry.RegisterSkill(entry))
	require.False(t, registry.ImplementsSkills())
	registry.RegisterResource(schema.Resource{Name: "demo", Uri: entry.Uri}, func(context.Context, *schema.ReadResourceRequest) (*schema.ReadResourceResult, *jsonrpc.Error) {
		return nil, nil
	})
	require.NoError(t, registry.RegisterSkill(entry))
	require.True(t, registry.ImplementsSkills())
}
