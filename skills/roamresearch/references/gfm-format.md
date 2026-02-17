# GFM-to-Roam Conversion Reference

`roam-cli save` converts GitHub Flavored Markdown (GFM) into Roam batch actions. This document describes the exact conversion rules.

## Conversion rules

| GFM element | How it maps to Roam |
|---|---|
| `#` (h1) | **Skipped** — page title is set via `--title`, do NOT include h1 in content |
| `##` – `###` | Block with `heading` attribute (2 or 3). Levels 4–6 are capped to 3 |
| `- item` / `* item` / `+ item` | Nested child block under current parent |
| Nested list (indented) | Deeper child block following indent level |
| `1. item` (ordered list) | Child block with text `1. item` (marker preserved) |
| `` ```lang ... ``` `` | Single block containing the full fenced code block |
| `> quote` | Single block with the `> ` prefix preserved |
| `---` / `***` / `___` | **Discarded** — horizontal rules are ignored |
| GFM table (with `|` and `---` separator) | Converted to `{{[[table]]}}` parent block with row/cell child blocks |
| Consecutive text lines | Joined into one paragraph block |

## Table format

Tables **must** be valid GFM tables with a header row, a separator row, and pipe delimiters. The converter will NOT recognise any other table format.

Correct example:

```markdown
| Name | Status |
| --- | --- |
| Alice | Done |
| Bob | WIP |
```

This produces:
- `{{[[table]]}}` parent block
  - Row 1 cells: `Name` → child `Status`
  - Row 2 cells: `Alice` → child `Done`
  - Row 3 cells: `Bob` → child `WIP`

**Do NOT** use HTML tables, indented tables, or any non-standard table format.

## What is NOT converted

- Images (`![alt](url)`) — passed through as-is in text
- Links (`[text](url)`) — passed through as-is in text
- Inline formatting (`**bold**`, `*italic**`, `` `code` ``) — passed through as-is
- Task lists (`- [ ] todo`) — treated as regular list items
