# roam-cli

`roam-cli` is a command-line tool for working with your Roam Research graph.

It supports:
- page/block retrieval
- full-text block search
- raw datalog queries
- markdown import as pages
- daily journal extraction
- low-level block and batch write operations

---

## Install

### Option A: Download from GitHub Releases

```bash
gh release list -R Leechael/roamresearch-skills
TAG="roam-cli-vX.Y.Z"
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
- `save` (`save-markdown`) — save markdown as a page
- `journal` (`get-journaling-by-date`, `journaling`) — read daily journaling blocks

### Output modes

- Parseable output is available via `--json`.
- Human-readable output is available via `--plain` (or default plain output when omitted).
- `--json` and `--plain` are mutually exclusive.
- `--jq` requires `--json` (supported on `status` and `q`).

### Low-level commands

- `block create|update|delete|get`
- `batch run`

---

## Usage examples

```bash
# status
roam-cli status --plain
roam-cli status --json
roam-cli status --json --jq '.ok'

# read
roam-cli get "Page Title" --plain
roam-cli get "((block-uid))" --json

# search
roam-cli search term1 term2 --limit 20 --plain
roam-cli search term1 term2 --limit 20 --json

# query
roam-cli q '[:find ?title :where [?e :node/title ?title]]' --json
roam-cli q '[:find ?title :where [?e :node/title ?title]]' --json --jq '.[0]'

# save markdown
roam-cli save --title "New Page" --file ./examples/note.md --json
cat ./examples/note.md | roam-cli save --title "New Page" --plain

# journal
roam-cli journal --date 2026-02-12 --plain
roam-cli journal --date 2026-02-12 --json

# low-level block
roam-cli block create --parent "02-12-2026" --text "hello" --json

# low-level batch
roam-cli batch run --file ./examples/actions.create-page-and-block.json --json
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
