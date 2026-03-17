# Usage Examples

## Read page/block

```bash
roam-cli get "Page Title"
roam-cli get "((block-uid))"
roam-cli get "Page Title" --json
```
## Search blocks

```bash
roam-cli search term1 term2 --limit 20
roam-cli search keyword --page "Project" --ignore-case

# Search on a daily page (pass ISO date, auto-resolved)
roam-cli search keyword --page 2026-03-14
```

## Datalog query

```bash
roam-cli q '[:find ?title :where [?e :node/title ?title]]'
```

## Save markdown

```bash
# Save to a new page
cat ./note.md | roam-cli save --title "New Page"
roam-cli save --title "New Page" --file ./note.md

# Save to today's daily page (one-shot)
cat ./note.md | roam-cli save --to-daily-page

# Save to a specific daily page
cat ./note.md | roam-cli save --to-daily-page 2026-03-14

# Save under an existing parent block
roam-cli save --parent <uid> --file ./note.md
```

## Journal by date

```bash
roam-cli journal --date 2026-02-12
roam-cli journal --date 2026-02-12 --topic "Work Log"
```

## Find block UID

```bash
# Find block by text on a daily note
roam-cli block find --text "[[📖 Daily Reading]]" --daily 2026-02-15

# Find block by text on a daily page (pass ISO date)
roam-cli block find --text "[[📖 Daily Reading]]" --page 2026-02-15

# Find block by text on a named page
roam-cli block find --text "Status" --page "Project Dashboard"
```

## Create blocks

`block create` supports single blocks, nested trees (JSON), and `--attach-to` for find-or-create.

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
roam-cli block create --parent <page-uid> --attach-to "[[📽 Journaling]]" --text "new item"
roam-cli block create --parent <page-uid> --attach-to "[[📽 Journaling]]" --file items.json
```

## Daily page workflows (one-shot)

```bash
# Save a tweet summary to today's daily page
echo '{"text":"🐦 tweet headline","children":[{"text":"key point"}]}' \
  | roam-cli save --to-daily-page

# Save article notes to a specific daily page
cat article.md | roam-cli save --to-daily-page 2026-03-14

# Insert under a section on a daily page (attach-to)
roam-cli block create --parent <daily-page-uid> --attach-to "[[📖 Daily Reading]]" --file article.json
```

## Low-level block operations

```bash
roam-cli block update --uid <uid> --text "updated"
roam-cli block move --uid <uid> --parent <target-parent-uid> --order last
roam-cli block delete --uid <uid>
roam-cli block get --uid <uid>
```

## Low-level batch operations

```bash
# Native batch actions
roam-cli batch run --file ./examples/actions.create-page-and-block.json
roam-cli batch run --file ./examples/actions.bulk-update.json
roam-cli batch run --file ./examples/actions.bulk-move.json

# Batch with children and attach-to
echo '[
  {"action":"create-block",
   "location":{"parent-uid":"page-uid","attach-to":"[[📽 Journaling]]","order":"last"},
   "block":{"string":"new item"}}
]' | roam-cli batch run
```
