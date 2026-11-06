package server

import (
	"context"
	"errors"
	"github.com/viant/mcp-protocol/schema"
	"testing"
)

func TestListToolsForCallerMetadata(t *testing.T) {
	registry := NewRegistry()
	denied := errors.New("denied")
	type roleKey struct{}
	registry.RegisterTool(&ToolEntry{Metadata: schema.Tool{Name: "private"}, MetadataFor: func(ctx context.Context) (schema.Tool, error) {
		if ctx.Value(roleKey{}) != "reader" {
			return schema.Tool{}, denied
		}
		return schema.Tool{Name: "public"}, nil
	}})
	if tools, err := registry.ListToolsFor(context.Background()); !errors.Is(err, denied) || len(tools) != 0 {
		t.Fatalf("leaked fallback: %v %v", tools, err)
	}
	tools, err := registry.ListToolsFor(context.WithValue(context.Background(), roleKey{}, "reader"))
	if err != nil || len(tools) != 1 || tools[0].Name != "public" {
		t.Fatalf("%v %v", tools, err)
	}
	if registry.ListRegisteredTools()[0].Name != "private" {
		t.Fatal("registration mutated")
	}
}
