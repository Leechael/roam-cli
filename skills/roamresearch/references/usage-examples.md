# Usage Examples

## Daily Use

### Read today's daily page

```bash
roam-cli get --today
roam-cli get --today --json
roam-cli get --daily yesterday
roam-cli get --daily 2026-03-14
```

### Read page/block

```bash
roam-cli get "Page Title"
roam-cli get "((block-uid))"
roam-cli get "Page Title" --json
```

### Journal by date

```bash
roam-cli journal --date today
roam-cli journal --date yesterday --topic "Work Log"
roam-cli journal --date 2026-02-12
```

### Search pages (default)

```bash
roam-cli search "meeting" "action item" --limit 10
roam-cli search "TODO" --daily-topic "[[TODO]]" --json

# Search on a daily page (pass ISO date, auto-resolved)
roam-cli search keyword --page 2026-03-14
```

### Search individual blocks

```bash
roam-cli search term1 term2 --type block --limit 20
roam-cli search keyword --type block --page "Project" -i
```

### Save to daily page (most common operation)

```bash
# Quick capture to today
printf '- meeting notes\n- action items' | roam-cli save --today

# Save under a section (find-or-create)
printf '- journal entry' | roam-cli save --today --under '[[📽 Journaling]]'

# Save reading highlights
cat highlights.md | roam-cli save --today --under '[[📖 Daily Reading]]'

# Save to a specific date
cat note.md | roam-cli save --to-daily-page 2026-03-14
```

### Create TODOs

```bash
printf '- {{[[TODO]]}} Review PR #12\n- {{[[TODO]]}} Call dentist' \
  | roam-cli save --today --under '[[TODO]]'
```

### Save to a named page

```bash
cat note.md | roam-cli save --title "New Page"
roam-cli save --title "Project X" --under '[[Tasks]]' --file ./tasks.md
```

### Compose: save then follow up

```bash
# Get target UID back
UID=$(printf '- headline' | roam-cli save --today --under '[[📽 Journaling]]' --plain)

# Add more under the same target
printf '- detail 1\n- detail 2' | roam-cli save --parent "$UID"
```

### Move blocks to organize

```bash
# Move a block to a project page section
roam-cli move --uid <block-uid> --title "Project X" --under '[[Tasks]]'

# Move to today's archive
roam-cli move --uid <block-uid> --today --under '[[Archive]]'

# Move to yesterday's daily page
roam-cli move --uid <block-uid> --daily yesterday
```

### Datalog query

```bash
roam-cli q '[:find ?title :where [?e :node/title ?title]]'
```

## Low-level API

### Find block UID

```bash
roam-cli block find --text "[[📖 Daily Reading]]" --today
roam-cli block find --text "Status" --page "Project Dashboard"
```

### Create blocks (JSON input)

```bash
# Single block
roam-cli block create --parent <uid> --text "hello"

# Nested tree from pipe (single object)
echo '{"text":"headline","children":[{"text":"snapshot"}]}' | roam-cli block create --parent <uid>

# Nested tree from pipe (array)
echo '[{"text":"item1"},{"text":"item2","children":[{"text":"sub"}]}]' | roam-cli block create --parent <uid>

# From file
roam-cli block create --parent <uid> --file ./tree.json

# Attach-to: find or create section, then insert under it
roam-cli block create --parent <uid> --attach-to "[[📽 Journaling]]" --text "new item"
roam-cli block create --parent <uid> --attach-to "[[📽 Journaling]]" --file items.json
```

### Block operations

```bash
roam-cli block update --uid <uid> --text "updated"
roam-cli block move --uid <uid> --parent <target-parent-uid> --order last
roam-cli block delete --uid <uid>
roam-cli block get --uid <uid>
```

### Batch operations

```bash
roam-cli batch run --file ./examples/actions.create-page-and-block.json
roam-cli batch run --file ./examples/actions.bulk-update.json

# Batch with children and attach-to
echo '[
  {"action":"create-block",
   "location":{"parent-uid":"page-uid","attach-to":"[[📽 Journaling]]","order":"last"},
   "block":{"string":"new item"}}
]' | roam-cli batch run
```
