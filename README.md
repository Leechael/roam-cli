# roam-cli

`roam-cli` is a command-line tool for working with your Roam Research graph.

It supports:
- page/block retrieval
- full-text block search
- raw datalog queries
- markdown import as pages or under daily pages
- daily journal extraction
- low-level block and batch write operations
- automatic batch chunking and rate-limit retry

---

## Install

### Option A: Download from GitHub Releases

```bash
gh release list -R Leechael/roamresearch-skills
TAG="vX.Y.Z"
./scripts/print-release-download.sh "$TAG"
# Example output:
# gh release download "$TAG" --pattern "roam-cli-*.tar.gz"
```

Extract the archive for your platform and place `roam-cli` in your `PATH`.

### Option B: Build from source

```bash
git clone git@github.com:Leechael/roamresearch-skills.git
cd roamresearch-skills
go build -o roam-cli ./cmd/roam-cli
```

---

## Required configuration

Set credentials via environment variables:

```bash
export ROAM_API_TOKEN="<token>"
export ROAM_API_GRAPH="<graph>"
```

Optional:

```bash
export ROAM_API_BASE_URL="https://api.roamresearch.com/api/graph"
export ROAM_TIMEOUT_SECONDS="10"
export TOPIC_NODE="<topic>"
```

Validate setup before use:

```bash
roam-cli status
roam-cli status --json
roam-cli status --json --jq '.ok'
```

---

## Commands

### High-level commands

- `status` — check credentials and API connectivity
- `get` — read page by title or block by uid
- `search` — search blocks by terms
- `q` — run raw datalog query
- `save` — save markdown as a page (`--title`), daily page (`--to-daily-page`), or under a parent block (`--parent`)
- `journal` — read daily journaling blocks
- `help` — show help or categorized examples (`help read`, `help write`, `help workflow`, `help all`)

### Low-level commands

- `block create|update|delete|move|get|find|create-tree`
- `batch run` (native actions + `create-with-children` DSL)

### Output modes

- `--json` — parseable JSON output
- `--plain` — human-readable plain text
- `--json` and `--plain` are mutually exclusive
- `--jq` — filter JSON (supported on `status` and `q`, requires `--json`)

### Pipeline support

Commands that accept input (`save`, `block create-tree`, `batch run`) read from stdin by default when `--file` is not given. No `--stdin` flag needed.

### Date handling

ISO dates (YYYY-MM-DD) are auto-resolved to Roam daily page titles wherever a page reference is expected:

- `save --to-daily-page 2026-03-14` → saves to "March 14th, 2026"
- `search --page 2026-03-14` → searches in "March 14th, 2026"
- `block find --page 2026-03-14` → finds in "March 14th, 2026"

---

## Usage examples

```bash
# status
roam-cli status --plain
roam-cli status --json --jq '.ok'

# read
roam-cli get "Page Title"
roam-cli get "((block-uid))" --json

# search
roam-cli search term1 term2 --limit 20
roam-cli search keyword --page 2026-03-14

# query
roam-cli q '[:find ?title :where [?e :node/title ?title]]' --json

# save markdown to a new page
cat note.md | roam-cli save --title "New Page"

# save to today's daily page
cat note.md | roam-cli save --to-daily-page

# save to a specific daily page
cat note.md | roam-cli save --to-daily-page 2026-03-14

# save under an existing parent block
roam-cli save --parent <uid> --file ./note.md

# journal
roam-cli journal --date 2026-02-12
roam-cli journal --date 2026-02-12 --topic "Work Log" --json

# find block on a daily page
roam-cli block find --text "[[📖 Daily Reading]]" --daily 2026-02-15
roam-cli block find --text "[[📖 Daily Reading]]" --page 2026-02-15

# create nested tree from JSON
echo '{"text":"headline","children":[{"text":"child"}]}' \
  | roam-cli block create-tree --parent <uid>
roam-cli block create-tree --parent <uid> --file ./tree.json

# low-level block operations
roam-cli block create --parent <uid> --text "hello"
roam-cli block update --uid <uid> --text "updated"
roam-cli block move --uid <uid> --parent <target-uid> --order last
roam-cli block delete --uid <uid>

# batch operations
roam-cli batch run --file ./examples/actions.create-page-and-block.json
echo '[...]' | roam-cli batch run

# categorized help
roam-cli help write
roam-cli help workflow
roam-cli help all
```

---

## Install the Agent Skill

This repository also ships an Agent Skill package under `skills/roamresearch`.

Install with:

```bash
npx skills add Leechael/roamresearch-skills
```

After installation, your agent can load and use the `roamresearch` skill instructions.

---

## Recommended secret handling

Use 1Password CLI to inject credentials at runtime:

- https://developer.1password.com/docs/service-accounts/use-with-1password-cli

Example:

```bash
op run --env-file=.env -- roam-cli status
op run --env-file=.env -- roam-cli get "Page Title"
```
