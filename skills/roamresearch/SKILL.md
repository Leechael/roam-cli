---
name: roamresearch
description: Daily Roam Research workflows via roam-cli. Use this skill for status checks, reading pages/blocks/daily pages, search, journaling lookup, markdown saves, moving blocks, and page clear/delete operations.
compatibility: Requires roam-cli in PATH and ROAM_API_TOKEN/ROAM_API_GRAPH environment variables.
---

# RoamResearch Skill

Use this skill for Daily Use workflows through `roam-cli`. Prefer named pages, daily-page flags, and `--under` sections. Do not use low-level `block`, `batch`, or `q` commands unless the user explicitly asks for low-level API work.

## Prerequisites

- `roam-cli` binary in PATH
- Environment variables: `ROAM_API_TOKEN`, `ROAM_API_GRAPH`
- Run `roam-cli status` before writes
- Local e2e tests may read `.env` from the current working directory; `.env` must stay gitignored

## Daily Use Commands

| Command | Use for |
|---|---|
| `status` | Verify credentials and Roam API reachability |
| `get` | Read a page by title, a block by UID, or a daily page |
| `search` | Search terms by page or by block |
| `journal` | Read daily journaling blocks |
| `save` | Save GFM markdown to a page, daily page, parent block, or section |
| `move` | Move an existing block to a page, daily page, parent, or section |
| `page clear` | Remove all top-level content from a page but keep the page |
| `page delete` | Delete a page |

## Read Workflows

```bash
roam-cli status
roam-cli get "Page Title"
roam-cli get "((block-uid))"
roam-cli get --today
roam-cli get --daily 2026-03-14
roam-cli journal --date today
```

Search defaults to page-level aggregation:

```bash
roam-cli search "term one" "term two" -i
roam-cli search "exact topic" --type page --limit 20
roam-cli search "needle" --type block --page "Page Title" --limit 10
```

Use `--type page` for broad retrieval. Use `--type block` when you need the exact matching block UID.

## Write Workflows

Use `save` for normal writes. It accepts GFM markdown from stdin or `--file`.

```bash
printf '%s\n' '- journal entry' | roam-cli save --today
printf '%s\n' '- item' | roam-cli save --today --under '[[Inbox]]'
cat note.md | roam-cli save --to-daily-page 2026-03-14
cat note.md | roam-cli save --title "Project Notes"
cat note.md | roam-cli save --title "Project Notes" --under '[[Tasks]]'
```

Use `--plain` when the next step needs the UID of the first created content block:

```bash
UID=$(printf '%s\n' '- parent item' | roam-cli save --today --under '[[Inbox]]' --plain)
printf '%s\n' '- child detail' | roam-cli save --parent "$UID"
```

Use `--replace` only when replacing a whole page. Do not combine it with `--under`.

```bash
cat note.md | roam-cli save --title "Project Notes" --replace
```

## Move Workflows

```bash
roam-cli move --uid BLOCK_UID --title "Project" --under '[[Tasks]]'
roam-cli move --uid BLOCK_UID --today --under '[[Archive]]'
roam-cli move --uid BLOCK_UID --parent PARENT_UID
```

Behavior verified by e2e: if `move --under` gets an invalid source UID, it fails before creating the destination section. This prevents empty sections from polluting the page.

## Page Operations

```bash
roam-cli page clear "Scratch Page"
roam-cli page delete "Scratch Page"
roam-cli page clear --daily 2026-03-14
roam-cli page delete --daily 2026-03-14
```

Behavior verified by e2e:

- `page clear` removes content and keeps the page readable.
- `page delete` deletes the page; a later `get` returns not found.

## `--under` Rules

`--under` finds a direct child block with exact matching text under the target page. If not found, it creates that section. It then appends the new content under that section.

Recommended:

```bash
printf '%s\n' '- content' | roam-cli save --today --under '[[📽 Journaling]]'
printf '%s\n' '- content' | roam-cli save --to-daily-page 2026-03-14 --under '[[📽 Journaling]]'
printf '%s\n' '- content' | roam-cli save --title "Project" --under '[[Notes]]'
```

Behavior verified by e2e: repeated `save --to-daily-page ... --under` on an existing daily page with existing nested content appends new content as a direct child of the section. It must not append into the previous item, a later top-level tail block, or any deeper child.

## Date Handling

Use ISO dates or relative dates. Do not manually construct Roam daily page titles.

| Task | Command |
|---|---|
| Read today | `roam-cli get --today` |
| Read a date | `roam-cli get --daily 2026-03-14` |
| Write today | `roam-cli save --today` |
| Write a date | `roam-cli save --to-daily-page 2026-03-14` |
| Move to today | `roam-cli move --uid BLOCK_UID --today --under '[[Section]]'` |
| Search one date | `roam-cli search "term" --page 2026-03-14` |

E2e note: Roam can reject `create-page` plus immediate `create-block` in the same batch for a future daily-title page. Daily workflows should use `--today` / `--to-daily-page` and avoid manually creating daily-title pages.

## Markdown Input Rules

- Prefer `printf | roam-cli save` over shell-heavy `--text` arguments.
- Do not add `--stdin` when piping; stdin is automatic.
- Do not include `# H1`; the page title comes from `--title` or daily-page flags.
- Lists become nested Roam blocks.
- Tables become `{{[[table]]}}` blocks.
- Code blocks and blockquotes are preserved.
- Horizontal rules are discarded.

## Anti-patterns

- Do not find a daily page UID before writing. Use `save --today` or `save --to-daily-page`.
- Do not find a section UID before writing. Use `save --under`.
- Do not manually construct daily page titles like `March 14th, 2026`; pass `2026-03-14`.
- Do not use `journal --json | jq` to get a parent UID for writes; `journal` returns journal blocks, not the page target.
- Do not use low-level commands for normal Daily Use flows.

## E2E Tests

Manual e2e tests live in `tests/e2e/daily_test.go` and are excluded from normal CI by the `e2e` build tag.

```bash
go test -tags e2e ./tests/e2e
go test -tags e2e ./tests/e2e -run 'TestDailyUse/SaveUnderExistingDailySectionKeepsSectionDepth'
go test -tags e2e ./tests/e2e -keep-pages
```

The Go e2e harness builds a fresh binary unless `ROAM_CLI` is set. Successful tests keep output quiet; failed subtests print the captured command log.
