---
name: roamresearch
description: Use for roam-cli Daily Use workflows: status, get/search/journal, save markdown to pages or daily sections, move blocks, and page clear/delete. Not for low-level block/batch/q API design, e2e maintenance, or Roam API internals.
compatibility: Requires roam-cli in PATH and ROAM_API_TOKEN/ROAM_API_GRAPH environment variables.
---

# RoamResearch Skill

Use `roam-cli` Daily Use commands. Prefer page titles, daily-page flags, and `--under` section names over manual UID lookup.

## Required First Step

Before writes, run:

```bash
roam-cli status
```

If credentials are missing, report the missing keys: `ROAM_API_TOKEN`, `ROAM_API_GRAPH`.

## Command Selection

| Need | Command |
|---|---|
| Verify setup | `roam-cli status` |
| Read page/block/daily page | `roam-cli get` |
| Search by page or block | `roam-cli search` |
| Read journaling blocks | `roam-cli journal` |
| Write markdown | `roam-cli save` |
| Move an existing block | `roam-cli move` |
| Empty a page | `roam-cli page clear` |
| Delete a page | `roam-cli page delete` |

## Read SOP

```bash
roam-cli get "Page Title"
roam-cli get "((block-uid))"
roam-cli get --today
roam-cli get --daily 2026-03-14
roam-cli journal --date today
```

Search defaults to page-level results. Use block mode when you need exact block hits.

```bash
roam-cli search "term one" "term two" -i
roam-cli search "topic" --type page --limit 20
roam-cli search "needle" --type block --page "Page Title" --limit 10
```

## Write SOP

Use `save` for normal writes. It reads stdin automatically.

```bash
printf '%s\n' '- journal entry' | roam-cli save --today
printf '%s\n' '- item' | roam-cli save --today --under '[[Inbox]]'
cat note.md | roam-cli save --to-daily-page 2026-03-14
cat note.md | roam-cli save --title "Project Notes"
cat note.md | roam-cli save --title "Project Notes" --under '[[Tasks]]'
```

Use `--plain` when the next command needs the first created block UID.

```bash
UID=$(printf '%s\n' '- parent item' | roam-cli save --today --under '[[Inbox]]' --plain)
printf '%s\n' '- child detail' | roam-cli save --parent "$UID"
```

Use `--replace` only for whole-page replacement.

```bash
cat note.md | roam-cli save --title "Project Notes" --replace
```

## `--under` Behavior

`--under` targets a direct child section under the target page. If the section exists, new content appends under that section. If it does not exist, `roam-cli` creates it first.

```bash
printf '%s\n' '- content' | roam-cli save --today --under '[[📽 Journaling]]'
printf '%s\n' '- content' | roam-cli save --to-daily-page 2026-03-14 --under '[[📽 Journaling]]'
printf '%s\n' '- content' | roam-cli save --title "Project" --under '[[Notes]]'
```

Repeated writes to the same daily-page section should remain direct children of that section, even when the page already has nested content.

## Move SOP

```bash
roam-cli move --uid BLOCK_UID --title "Project" --under '[[Tasks]]'
roam-cli move --uid BLOCK_UID --today --under '[[Archive]]'
roam-cli move --uid BLOCK_UID --parent PARENT_UID
```

If the source UID is invalid, `move --under` fails before creating the target section.

## Page SOP

```bash
roam-cli page clear "Scratch Page"
roam-cli page delete "Scratch Page"
roam-cli page clear --daily 2026-03-14
roam-cli page delete --daily 2026-03-14
```

`page clear` keeps the page. `page delete` removes it; later `get` should return not found.

## Date SOP

Pass ISO or relative dates. Let `roam-cli` convert to Roam daily page titles.

| Task | Command |
|---|---|
| Read today | `roam-cli get --today` |
| Read date | `roam-cli get --daily 2026-03-14` |
| Write today | `roam-cli save --today` |
| Write date | `roam-cli save --to-daily-page 2026-03-14` |
| Move to today | `roam-cli move --uid BLOCK_UID --today --under '[[Section]]'` |
| Search date | `roam-cli search "term" --page 2026-03-14` |

## Markdown Input

- Prefer `printf | roam-cli save` for content containing `[[refs]]`, emoji, quotes, or shell-sensitive text.
- Omit `--stdin` when piping; stdin is automatic.
- Omit `# H1`; page title comes from flags.
- Lists become nested blocks.
- Tables become Roam table blocks.
- Code blocks and blockquotes are preserved.
- Horizontal rules are discarded.

## Default Path

1. Verify: `roam-cli status`.
2. Read with `get`, `search`, or `journal`.
3. Write with `save` using `--today`, `--to-daily-page`, `--title`, `--under`, or `--parent`.
4. Organize with `move`.
5. Clean pages with `page clear` or `page delete`.

## Help Topics

Read built-in help before unfamiliar workflows:

```bash
roam-cli help writing-guide
roam-cli help write-examples
roam-cli help workflow-examples
roam-cli help read-examples
```
