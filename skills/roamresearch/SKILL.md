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
| `search-pages` | Multi-query search aggregated by page, with daily page drill-down |
| `q` | Run raw datalog query |
| `save` | Save GFM markdown as a page, daily page (`--to-daily-page` / `--today`), or under a parent block; `--under` finds-or-creates a child block |
| `journal` | Read daily journaling blocks |
| `block find` | Find block UID by text on a page/daily note |
| `block create` | Create block(s) — single, nested tree, or with `--attach-to` |
| `block update/delete/move/get` | Low-level single-block operations |
| `batch run` | Batch actions from JSON array |

Run `roam-cli help <category>` for categorized usage examples. Categories: `read`, `write`, `workflow`, or `all`.

## Search Strategy

**Use `search-pages` for broad retrieval, `search` for targeted lookup.**

| Scenario | Use this |
|---|---|
| Find pages related to a topic across the whole graph | `search-pages` with multiple query args |
| Find specific blocks containing exact terms | `search` |
| Search within a known page | `search --page "Page Title"` |

### `search-pages` — multi-query page-level search

Each positional argument is an independent query. Terms within one argument are AND-matched (block must contain all terms). Results are deduplicated, aggregated by page, and sorted by queries matched then hit count.

```bash
# Multiple independent queries — results merged and ranked
roam-cli search-pages "SaaS 倒闭" "AI 贷款" "SaaS shutdown" -i

# With --json for structured output (pipe to rerank tools)
roam-cli search-pages "query1" "query2" -i --json

# Limit results
roam-cli search-pages "broad term" -i --limit 20
```

**Daily page handling:** Daily pages (e.g. "March 8th, 2026") are automatically drilled down to show sections instead of the whole page.

```bash
# Default: aggregate at first-level children of daily pages
roam-cli search-pages "AI 贷款" -i

# Drill deeper: aggregate at second level, filter first level by topic
roam-cli search-pages "AI 贷款" -i --daily-depth 2 --daily-topic "📖 Daily Reading"
```

| Flag | Purpose |
|---|---|
| `--daily-depth N` | Aggregation depth for daily pages (default 1) |
| `--daily-topic TEXT` | Filter daily page sections at level 1 by topic text |
| `--limit N` | Max results to return (0 = unlimited) |
| `-i` | Case-insensitive search |
| `--json` | Structured JSON output |

Output includes `((block_uid))` for each result — use `roam-cli get "((uid))"` to retrieve full content with children.

## Write Strategy (critical — read this first)

**Minimize API calls.** Every tool call costs tokens. Use the highest-level command that fits.

| Scenario | Use this | NOT this |
|---|---|---|
| Save content to today's daily page | `save --today` | `journal` → parse UID → `save --parent` |
| Save content to a specific daily page | `save --to-daily-page 2026-03-14` | `block find --daily` → `save --parent` |
| Save under a section in today's daily page | `save --today --under '[[Section]]'` | `block find` → `block create --parent` |
| Save under a section in any page | `save --title "Page" --under '[[Section]]'` | `block find` → `block create --parent` |
| Save a long document/article as a page | `save --title "Page Name"` | Sequential `block create` |
| Create a parent with children | `block create --parent <uid> --file tree.json` | `block create` parent → `block create` child × N |
| Insert under existing section | `block create --parent <uid> --attach-to "[[Section]]"` | `block find` → `block create --parent` |
| Multiple heterogeneous writes | `batch run` | Multiple individual write calls |
| Single block, no children | `block create --parent <uid> --text "foo"` | (this is fine) |

### `block create` modes

```bash
# Single block
roam-cli block create --parent <uid> --text "Hello"

# Nested tree (JSON input via file or stdin)
echo '{"text":"Root","children":[{"text":"Child"}]}' | roam-cli block create --parent <uid>
roam-cli block create --parent <uid> --file tree.json

# Attach-to: find or create a section block, then insert under it
roam-cli block create --parent <page-uid> --attach-to "[[📽 Journaling]]" --text "new item"
roam-cli block create --parent <page-uid> --attach-to "[[📽 Journaling]]" --file items.json
```

`--attach-to` finds an existing block with matching text under `--parent`. If not found, creates it first. Then creates the content under that block.

### Daily page operations

Use `--today` or `--to-daily-page` for one-shot writes to daily pages. Do NOT manually construct Roam daily page titles like "March 14th, 2026" — the CLI handles this internally. Pages are automatically upserted (created if missing, appended to if existing).

```bash
# Save markdown to today's daily page
echo "- entry" | roam-cli save --today

# Save to a specific date
cat note.md | roam-cli save --to-daily-page 2026-03-14

# Save under a section in today's daily page (find-or-create)
echo "- journal entry" | roam-cli save --today --under '[[📽 Journaling]]'

# Save under a section in a named page
cat note.md | roam-cli save --title "Project Notes" --under '[[Tasks]]'

# Search/find on a daily page — pass ISO date, CLI auto-resolves
roam-cli search --page 2026-03-14 keyword
roam-cli block find --page 2026-03-14 --text "[[📖 Daily Reading]]"
```

`--under` finds an existing direct child block with matching text under the target page. If not found, creates it first. Then appends content under that block. This is the recommended way to write to daily page sections like `[[📽 Journaling]]`.

### Anti-patterns — do NOT do these

- Do NOT call `block create` in a loop to build a tree. Use JSON input with children.
- Do NOT fire multiple `block create` in parallel to the same parent. Use `batch run` or JSON tree input.
- Do NOT do multi-step "find block → then create under it". Use `--attach-to`.
- Do NOT do multi-step "find daily page UID → then write". Use `save --today` or `save --to-daily-page`.
- Do NOT do multi-step "find daily page → find section → write under it". Use `save --today --under '[[Section]]'`.
- Do NOT manually construct "Month DDth, YYYY" date strings. Pass ISO dates (YYYY-MM-DD) to `--to-daily-page`, `--daily`, or `--page` — the CLI converts them.
- Do NOT add `--stdin` when piping — it's automatic.

## Pipeline Support

All commands that accept input (`save`, `block create`, `batch run`) read from stdin by default when `--file` is not given. No `--stdin` flag needed.

```bash
echo '{"text":"root","children":[{"text":"child"}]}' | roam-cli block create --parent <uid>
cat note.md | roam-cli save --title "Page Name"
echo '[...]' | roam-cli batch run
```

## `block create` JSON Input Contract

- Requires `--parent <block-uid>`.
- Accepts JSON from pipe or `--file`.
- JSON supports either a single object or an array of objects.
- Node shape: `text` (required), `children` (optional array of nodes).
- Both `text` and `string` keys accepted (`text` takes precedence).

```json
{"text": "headline", "children": [
  {"text": "point 1"},
  {"text": "point 2", "children": [{"text": "sub-point"}]}
]}
```

## `batch run` Actions

**Native actions** (pass-through to Roam API):
- `create-block` — supports `children` in block field (auto-expanded) and `attach-to` in location
- `update-block` — requires `block.uid` and `block.string`
- `delete-block` — requires `block.uid`
- `move-block` — requires `block.uid` and `location.parent-uid`
- `create-page` — requires `page.title`

```json
[
  {"action": "create-block",
   "location": {"parent-uid": "PAGE_UID", "attach-to": "[[📽 Journaling]]", "order": "last"},
   "block": {"string": "new item under Journaling"}},

  {"action": "create-block",
   "location": {"parent-uid": "PARENT_UID", "order": "last"},
   "block": {"string": "Parent", "children": [
     {"string": "Child 1"},
     {"string": "Child 2", "children": [{"string": "Grandchild"}]}
   ]}}
]
```

## Date Handling

The CLI auto-resolves ISO dates (YYYY-MM-DD) to Roam daily page titles wherever a page reference is expected:

| Flag | Input | Resolved to |
|---|---|---|
| `save --to-daily-page` / `--today` | `2026-03-14` / today | Creates/finds page "March 14th, 2026" |
| `search --page` | `2026-03-14` | Searches in "March 14th, 2026" |
| `block find --page` | `2026-03-14` | Finds block in "March 14th, 2026" |
| `block find --daily` | `2026-03-14` | Finds by daily page UID (existing behavior) |
| `journal --date` | `2026-03-14` | Reads daily journal (existing behavior) |

## Recommended Workflow

1. `roam-cli status` — verify credentials.
2. Read: `get`, `search`, `search-pages`, `q`, `journal`, `block find`.
3. Write (pick one, in order of preference):
   - Daily page content → `save --today` or `save --to-daily-page`
   - Daily page section → `save --today --under '[[Section]]'`
   - Long markdown → `save --title`
   - Nested blocks / attach to existing section → `block create` (with JSON + `--attach-to`)
   - Mixed operations → `batch run`
   - Single block → `block create --text`

## Save Markdown (GFM format)

`save` accepts GFM and auto-converts to Roam blocks:

- Do NOT include `#` h1 — title comes from `--title` or `--to-daily-page`
- `##`–`###` → headed blocks (levels 4–6 capped to 3)
- Lists → nested child blocks; ordered lists preserve marker
- Tables → `{{[[table]]}}` blocks (must be valid GFM pipe+separator)
- Code blocks, blockquotes → passed through
- Horizontal rules → discarded

Full rules: `references/gfm-format.md`

## Error Handling Rules

- Missing credentials: report missing `ROAM_API_TOKEN` / `ROAM_API_GRAPH`.
- API failures: include HTTP status code and response body.
- Not found: include the identifier/uid that was requested.

## Output Rules

- Preserve JSON output when `--json` is requested.
- Keep default output concise and readable.
- Never invent Roam data; only report real command results.

## Detailed Examples

Run `roam-cli help all` or see `references/usage-examples.md`.
