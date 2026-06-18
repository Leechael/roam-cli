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

func TestFormatBlocksAsMarkdown_DoesNotIndentNestedHeadings(t *testing.T) {
	blocks := []map[string]any{
		{
			":block/string": "Parent",
			":block/children": []any{
				map[string]any{
					":block/string":  "Nested Heading",
					":block/heading": 2,
					":block/order":   0,
				},
			},
		},
	}
	out := FormatBlocksAsMarkdown(blocks)
	if !strings.Contains(out, "\n## Nested Heading") {
		t.Fatalf("expected nested heading without indentation, got: %q", out)
	}
	if strings.Contains(out, "\n  ## Nested Heading") {
		t.Fatalf("expected nested heading not to be indented, got: %q", out)
	}
}

func TestFormatBlocksAsMarkdown_DoesNotIndentNestedFencedCode(t *testing.T) {
	blocks := []map[string]any{
		{
			":block/string": "Parent",
			":block/children": []any{
				map[string]any{
					":block/string": "Child",
					":block/order":  0,
					":block/children": []any{
						map[string]any{
							":block/string": "```go\nfmt.Println(\"hi\")\n```",
							":block/order":  0,
						},
					},
				},
			},
		},
	}
	out := FormatBlocksAsMarkdown(blocks)
	if !strings.Contains(out, "\n```go\nfmt.Println(\"hi\")\n```") {
		t.Fatalf("expected nested fenced code without indentation, got: %q", out)
	}
	if strings.Contains(out, "\n    ```go") {
		t.Fatalf("expected nested fenced code not to be indented, got: %q", out)
	}
}
