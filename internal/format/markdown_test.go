package format

import (
	"strings"
	"testing"
)

func TestFormatBlocksAsMarkdown(t *testing.T) {
	blocks := []map[string]any{
		{
			":block/string":  "Title",
			":block/heading": 2,
			":block/children": []any{
				map[string]any{":block/string": "child line", ":block/order": 0},
			},
		},
	}
	out := FormatBlocksAsMarkdown(blocks)
	if out == "" {
		t.Fatal("expected non-empty markdown")
	}
	if out[:2] != "##" {
		t.Fatalf("expected heading markdown, got: %s", out)
	}
}

func TestFormatBlocksAsMarkdown_IndentHierarchy(t *testing.T) {
	blocks := []map[string]any{
		{
			":block/string": "Parent",
			":block/children": []any{
				map[string]any{
					":block/string": "Child",
					":block/order":  0,
					":block/children": []any{
						map[string]any{":block/string": "Grandchild", ":block/order": 0},
					},
				},
			},
		},
	}
	out := FormatBlocksAsMarkdown(blocks)
	if !strings.Contains(out, "\n  Child") {
		t.Fatalf("expected child indentation, got: %q", out)
	}
	if !strings.Contains(out, "\n    Grandchild") {
		t.Fatalf("expected grandchild indentation, got: %q", out)
	}
}
