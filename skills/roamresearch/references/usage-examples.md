# Usage Examples

## Read page/block

```bash
roam-cli get "Page Title"
roam-cli get "((block-uid))"
roam-cli get "Page Title" --raw
```

## Search blocks

```bash
roam-cli search term1 term2 --limit 20
roam-cli search keyword --page "Project" --ignore-case
```

## Datalog query

```bash
roam-cli q '[:find ?title :where [?e :node/title ?title]]'
```

## Save markdown

```bash
roam-cli save --title "New Page" --file ./note.md
cat ./note.md | roam-cli save --title "New Page"
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

# Find block by text on a named page
roam-cli block find --text "Status" --page "Project Dashboard"
```

## Create nested block tree

```bash
# From stdin (single object)
echo '{"text":"headline","children":[{"text":"snapshot"}]}' | roam-cli block create-tree --parent <uid> --stdin

# From stdin (array)
echo '[{"text":"item1"},{"text":"item2","children":[{"text":"sub"}]}]' | roam-cli block create-tree --parent <uid> --stdin

# From file
roam-cli block create-tree --parent <uid> --file ./tree.json
```

## Optimized daily-note workflow (2 calls)

```bash
# Step 1: Find the target block UID
UID=$(roam-cli block find --daily 2026-02-15 --text "[[📖 Daily Reading]]")

# Step 2: Create nested tree under it
echo '{"text":"headline","children":[{"text":"snapshot"}]}' | roam-cli block create-tree --parent "$UID" --stdin
```

## Low-level block operations

```bash
roam-cli block create --parent <uid> --text "hello"
roam-cli block update --uid <uid> --text "updated"
roam-cli block delete --uid <uid>
roam-cli block get --uid <uid>
```

## Low-level batch operations

```bash
roam-cli batch run --file ./examples/actions.create-page-and-block.json
cat ./examples/actions.create-page-and-block.json | roam-cli batch run --stdin
```
