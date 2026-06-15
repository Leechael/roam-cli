# Writing Guide

Choose the right command for writing content to Roam Research.

## Overview

roam-cli has two layers of write commands:

- **save** -- high-level, takes markdown, resolves targets by name/date. Use this
  for daily notes, journaling, article saves, and most content creation.
  Add `--replace` when you want to overwrite an existing page.
- **page clear / page delete** -- clear a page's content or delete a page by
  title/date in one step.
- **block create** -- low-level, takes JSON trees or single text, requires explicit
  parent UID. Use this for batch imports, precise block placement, or when you
  need control over UIDs and ordering.

If you are unsure which to use, start with `save`.

## Commands

### save (recommended for most writes)

```bash
# Append to today's daily page
printf '%s\n' '- meeting notes' '- action items' | roam-cli save --today

# Append under a section on today's daily page
printf '%s\n' '- journal entry' | roam-cli save --today --under '[[📽 Journaling]]'

# Save to a named page (creates if missing)
cat article.md | roam-cli save --title "Article: How Roam Works"

# Save under a section on a named page
printf '%s\n' '- new task' | roam-cli save --title "Project X" --under '[[TODO]]'

# Get the UID back for follow-up commands
UID=$(printf '%s\n' '- item' | roam-cli save --today --under '[[📽 Journaling]]' --plain)
```

### save with replacement

```bash
# Replace a named page
cat note.md | roam-cli save --title "Project X" --replace

# Replace today's daily page
cat note.md | roam-cli save --today --replace
```

### page (clear or delete)

```bash
# Clear all content from a page
roam-cli page clear "Project X"

# Delete a page
roam-cli page delete "Project X"

# Clear today's daily page
roam-cli page clear --today
```

### block create (low-level)

```bash
# Single block under a known parent UID
roam-cli block create --parent UID_HERE --text "hello"

# JSON tree from stdin
echo '{"text":"Root","children":[{"text":"Child"}]}' | roam-cli block create --parent UID_HERE

# Find-or-create section, then insert JSON under it
roam-cli block create --parent UID_HERE --attach-to "[[Section]]" --file items.json
```

### Creating TODOs

Roam uses `{{[[TODO]]}}` syntax for tasks. Use printf to pipe the content:

```bash
# Single TODO
printf '%s\n' '- {{[[TODO]]}} Review PR #12' | roam-cli save --today --under '[[TODO]]'

# Multiple TODOs
printf '%s\n' '- {{[[TODO]]}} Buy groceries' '- {{[[TODO]]}} Call dentist' | roam-cli save --today --under '[[TODO]]'
```

### Composing commands

save --plain outputs the target UID, which can feed into subsequent commands:

```bash
# Save content, then add more under the same target
UID=$(printf '%s\n' '- headline' | roam-cli save --today --under '[[📽 Journaling]]' --plain)
printf '%s\n' '- detail 1' '- detail 2' | roam-cli save --parent "$UID"

# Save, then move another block under the same section
roam-cli move --uid <existing-block> --today --under '[[📽 Journaling]]'
```

## Constraints

- `save` only accepts markdown input (stdin or --file). For JSON tree input,
  use `block create`.
- `save --under` and `block create --attach-to` do the same thing (find-or-create
  a child block). `--under` is the high-level name, `--attach-to` is the low-level
  name.
- `save` to a page appends by default. Use `--replace` to clear the existing
  page content first.
- Use `page clear` to clear page content without deleting the page, or
  `page delete` to remove the page itself.
- `block create --parent` requires a UID. If you do not have one, use `save` with
  `--title`, `--today`, or `--to-daily-page` instead.
- `save --plain` outputs the target UID (page UID in page mode, parent UID in
  parent mode). This is the UID content was written under, not individual block UIDs.

## Anti-patterns

Do NOT manually look up page UIDs to pass to block create:

```bash
# WRONG: journal returns block UIDs, not page UIDs
PAGE_UID=$(roam-cli journal --json | jq -r '.[0][":block/uid"]')
roam-cli block create --parent "$PAGE_UID" --attach-to "[[Section]]" --text "item"

# RIGHT: save resolves the daily page internally
printf '%s\n' '- item' | roam-cli save --today --under '[[Section]]'
```

Do NOT use block create when save would work:

```bash
# WRONG: unnecessary complexity
roam-cli block create --parent UID_HERE --text "$(cat note.md)"

# RIGHT: save handles markdown properly
cat note.md | roam-cli save --parent UID_HERE
```

Prefer printf piping over --text for content with special characters:

```bash
# FRAGILE: shell may eat [[ ]] or emoji
roam-cli block create --parent UID_HERE --text "[[📽 Journaling]] entry"

# ROBUST: printf preserves content
printf '%s\n' '- [[📽 Journaling]] entry' | roam-cli save --today
```

## Examples

Daily journaling:
```bash
printf '%s\n' '- Started working on feature X' '- Met with team about deadlines' \
  | roam-cli save --today --under '[[📽 Journaling]]'
```

Save reading highlights:
```bash
cat highlights.md | roam-cli save --today --under '[[📖 Daily Reading]]'
```

Quick capture to inbox:
```bash
printf '%s\n' '- Check out this article: https://example.com' \
  | roam-cli save --today --under '[[Inbox]]'
```

Batch import from a file:
```bash
roam-cli save --title "Meeting Notes 2026-03-29" --file ./meeting.md
```

## Related Topics

- `roam-cli help exit-codes`
- `roam-cli help datalog`
