# GFM Format

How `save` converts GitHub Flavored Markdown (GFM) into Roam blocks.

## Overview

The `save` command accepts GFM markdown from stdin or `--file` and converts it
into Roam batch actions. Each markdown element maps to a specific Roam block
structure. Inline formatting (bold, italic, code, links, images) is passed
through as-is -- Roam renders standard markdown inline syntax natively.

## Conversion Rules

| GFM element | Roam result |
|---|---|
| `#` (h1) | Skipped -- page title comes from `--title`, do not include h1 |
| `##` -- `###` | Block with heading attribute (levels 4--6 capped to 3) |
| `- item` / `* item` / `+ item` | Nested child block |
| Indented list items | Deeper child blocks following indent level |
| `1. item` (ordered list) | Child block with marker preserved: `1. item` |
| `` ```lang ... ``` `` | Single block containing the full fenced code block |
| `> quote` | Single block with `> ` prefix preserved |
| `---` / `***` / `___` | Discarded (horizontal rules ignored) |
| GFM table (pipe + separator) | `{{[[table]]}}` parent with row/cell children |
| Consecutive text lines | Joined into one paragraph block |

## Commands

- `roam-cli save --today` -- save GFM to today's daily page
- `roam-cli save --title "Page"` -- save GFM as a named page
- `roam-cli save --today --under '[[Section]]'` -- save under a section

## Constraints

- Tables must be valid GFM: header row, separator row (`| --- | --- |`), and
  pipe delimiters. HTML tables and non-standard formats are not recognized.
- `#` (h1) is always skipped. Use `--title` for the page name.
- Heading levels 4--6 are capped to level 3 in Roam.
- Task list syntax (`- [ ] todo`) is treated as a regular list item. Use
  `{{[[TODO]]}}` for Roam-native tasks.
- Images and links are passed through as text -- Roam renders them natively.

## Examples

Save a structured document:

```bash
cat <<'EOF' | roam-cli save --title "Weekly Report"
## Highlights
- Feature A shipped
- Bug B fixed

## Metrics
| Metric | Value |
| --- | --- |
| PRs merged | 12 |
| Issues closed | 8 |

## Code snippet
` `` go
func main() { fmt.Println("hello") }
` ``
EOF
```

Save a list to a daily page section:

```bash
printf '- Meeting with Alice\n  - Discussed roadmap\n  - Action: review spec\n- Lunch break' \
  | roam-cli save --today --under '[[Meeting Notes]]'
```

## Related Topics

- `roam-cli help writing-guide`
- `roam-cli help exit-codes`
