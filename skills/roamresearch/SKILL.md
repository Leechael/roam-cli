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

Commands are grouped into two tiers. **Use Daily Use commands first** — they handle page resolution internally. Only fall back to Low-level API when you need explicit UID control or JSON input.

### Daily Use (markdown, name-based targeting)

| Command | Purpose |
|---|---|
| `save` | Write GFM markdown to a page, daily page, or section. Converts headings, lists, tables, code blocks. **Default choice for all writes.** |
| `get` | Read page by title, block by UID, or daily page (`--today` / `--daily`) |
| `search` | Search by terms. `--type page` (default) aggregates by page; `--type block` returns individual blocks |
| `journal` | Read daily journaling blocks |
| `move` | Move a block to a page or section by name (`--title` / `--today` / `--under`) |
| `status` | Verify credentials and API connectivity |

### Low-level API (JSON, explicit UIDs)

| Command | Purpose |
|---|---|
| `block create` | Create block(s) from JSON tree or single text under a known parent UID |
| `block update/delete/move/get/find` | Single-block CRUD by UID |
| `batch run` | Batch actions from JSON array |
| `q` | Run raw datalog query |

Run `roam-cli help <category>` for categorized usage examples. Categories: `read`, `write`, `workflow`, or `all`.

## Search Strategy

**Use `search` (default `--type page`) for broad retrieval, `search --type block` for targeted block lookup.**

| Scenario | Use this |
|---|---|
| Find pages related to a topic across the whole graph | `search` with multiple query args (defaults to `--type page`) |
| Find specific blocks containing exact terms | `search` |
| Search within a known page | `search --page "Page Title"` |

### `search` — page-level search (default mode)

Each positional argument is an independent query. Terms within one argument are AND-matched (block must contain all terms). Results are deduplicated, aggregated by page, and sorted by queries matched then hit count.

```bash
# Multiple independent queries — results merged and ranked
roam-cli search "SaaS 倒闭" "AI 贷款" "SaaS shutdown" -i

# With --json for structured output (pipe to rerank tools)
roam-cli search "query1" "query2" -i --json

# Limit results
roam-cli search "broad term" -i --limit 20
```

**Daily page handling:** Daily pages (e.g. "March 8th, 2026") are automatically drilled down to show sections instead of the whole page.

```bash
# Default: aggregate at first-level children of daily pages
roam-cli search "AI 贷款" -i

# Drill deeper: aggregate at second level, filter first level by topic
roam-cli search "AI 贷款" -i --daily-depth 2 --daily-topic "📖 Daily Reading"
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
| Move block to a project page | `move --uid <uid> --title "Project" --under '[[Tasks]]'` | `block find` → `block move --parent` |
| Move block to today's section | `move --uid <uid> --today --under '[[Archive]]'` | Manual UID lookup → `block move` |
| Create a parent with children (JSON) | `block create --parent <uid> --file tree.json` | `block create` parent → `block create` child × N |
| Insert JSON under existing section | `block create --parent <uid> --attach-to "[[Section]]"` | `block find` → `block create --parent` |
| Multiple heterogeneous writes | `batch run` | Multiple individual write calls |
| Single block, no children | `block create --parent <uid> --text "foo"` | (this is fine) |

**Prefer `printf | save` over constructing `--text` arguments.** Shell escaping with `[[references]]` and emoji is fragile:

```bash
# Recommended
printf '- {{[[TODO]]}} Review PR\n- entry with [[📽 Journaling]]' | roam-cli save --today --under '[[TODO]]'

# Fragile — shell may eat [[ ]] or emoji
roam-cli block create --parent <uid> --text "[[📽 Journaling]] entry"
```

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
- Do NOT do multi-step "find block → then create under it". Use `save --under` or `block create --attach-to`.
- Do NOT do multi-step "find daily page UID → then write". Use `save --today` or `save --to-daily-page`.
- Do NOT do multi-step "find daily page → find section → write under it". Use `save --today --under '[[Section]]'`.
- Do NOT use `journal --json | jq` to extract a UID and pass to `block create --parent`. Journal returns **block UIDs**, not page UIDs. Use `save --today --under` instead.
- Do NOT use `block create` when `save` would work. `save` handles markdown, page resolution, and upsert internally.
- Do NOT manually construct "Month DDth, YYYY" date strings. Pass ISO dates (YYYY-MM-DD) or relative dates (`today`, `yesterday`, `tomorrow`) — the CLI converts them.
- Do NOT add `--stdin` when piping — it's automatic.

## Pipeline Support

All commands that accept input (`save`, `block create`, `batch run`) read from stdin by default when `--file` is not given. No `--stdin` flag needed.

```bash
printf '- journal entry' | roam-cli save --today --under '[[📽 Journaling]]'
cat note.md | roam-cli save --title "Page Name"
echo '{"text":"root","children":[{"text":"child"}]}' | roam-cli block create --parent <uid>
echo '[...]' | roam-cli batch run
```

### Composing commands with `--plain`

`save --plain` outputs the target UID (page or parent block) for follow-up commands:

```bash
# Save and get UID back
UID=$(printf '- item' | roam-cli save --today --under '[[Inbox]]' --plain)

# Add more content under the same target
printf '- detail' | roam-cli save --parent "$UID"

# Or move another block there
roam-cli move --uid <existing-block> --today --under '[[Inbox]]'
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

The CLI auto-resolves dates to Roam daily page titles. Accepts ISO dates (YYYY-MM-DD) or relative dates (`today`, `yesterday`, `tomorrow`).

| Flag | Input | Resolved to |
|---|---|---|
| `save --to-daily-page` / `--today` | `2026-03-14` / `today` | Creates/finds page "March 14th, 2026" |
| `get --today` / `--daily` | `today` / `yesterday` / `2026-03-14` | Reads daily page |
| `move --today` / `--daily` | `today` / `2026-03-14` | Moves block to daily page |
| `search --page` | `2026-03-14` | Searches in "March 14th, 2026" |
| `journal --date` | `today` / `yesterday` / `2026-03-14` | Reads daily journal |
| `block find --today` / `--daily` | `today` / `2026-03-14` | Finds block by daily page |

## Recommended Workflow

1. `roam-cli status` — verify credentials.
2. Read:
   - Daily page → `get --today` or `get --daily yesterday`
   - Page/block → `get "Page Title"` or `get "((uid))"`
   - Journal → `journal --date today`
   - Search → `search` (`--type page` default, `--type block` for blocks)
3. Write (pick one, in order of preference):
   - Daily page section → `printf '...' | save --today --under '[[Section]]'`
   - Daily page content → `save --today`
   - Named page → `save --title "Page Name"`
   - Nested JSON blocks → `block create --parent <uid> --file tree.json`
   - Mixed operations → `batch run`
4. Organize:
   - Move to named page → `move --uid <uid> --title "Page" --under '[[Section]]'`
   - Move to daily page → `move --uid <uid> --today --under '[[Section]]'`

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
