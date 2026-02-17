---
name: roamresearch
description: Roam Research operations via roam-cli. Use this skill for page/block retrieval, search, datalog queries, markdown save, journaling lookup, and low-level block or batch writes with environment-injected credentials.
compatibility: Requires roam-cli in PATH and ROAM_API_TOKEN/ROAM_API_GRAPH environment variables.
---

# RoamResearch Skill

Use this skill to perform Roam Research read/write/query workflows through `roam-cli`.

## Prerequisites

- `roam-cli` binary in PATH
- Environment variables: `ROAM_API_TOKEN`, `ROAM_API_GRAPH`
- Run `roam-cli status` to verify before any operations

If not set up, see `references/installation.md`.

## Command Mapping

| Command | Purpose |
|---|---|
| `get` | Read page by title or block by UID |
| `search` | Search blocks by terms |
| `q` | Run raw datalog query |
| `save` | Save GFM markdown as a page |
| `journal` | Read daily journaling blocks |
| `block find` | Find block UID by text on a page/daily note |
| `block create-tree` | Create nested block tree from JSON |
| `block create/update/delete/get` | Low-level block CRUD |
| `batch run` | Low-level batch actions |

## Recommended Workflow

1. Run `roam-cli status` first.
2. Prefer read commands first: `get`, `search`, `q`.
3. Prefer high-level writes: `save`, `journal`.
4. Use low-level APIs for deterministic control: `block`, `batch`.

## Save Markdown (GFM format)

`save` accepts **GFM (GitHub Flavored Markdown)** and auto-converts to Roam blocks. Key rules:

- Do NOT include `#` h1 — page title comes from `--title`
- Headings `##`–`###` become headed blocks (levels 4–6 capped to 3)
- Lists (`-`/`*`/`+`) become nested child blocks; ordered lists preserve the `1.` marker
- Tables MUST use standard GFM pipe+separator format (header row, `---` row, data rows)
- Code blocks, blockquotes passed through; horizontal rules discarded

Full conversion rules: see `references/gfm-format.md`

## Error Handling Rules

- Missing credentials: explicitly report missing `ROAM_API_TOKEN` / `ROAM_API_GRAPH`.
- API failures: include HTTP status code and response body.
- Not found: clearly include the identifier/uid that was requested.

## Output Rules

- Preserve JSON output when `--raw` is requested.
- Keep default output concise and readable.
- Never invent Roam data; only report real command results.

## Detailed Examples

See `references/usage-examples.md` for full command examples.
