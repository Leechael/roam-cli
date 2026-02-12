# roamresearch-skills / roam-cli

Go CLI refactor workspace for the RoamResearch MCP/SDK.

## Current status

Core commands are implemented:
- high-level:
  - `get`
  - `search`
  - `q`
  - `save` (alias: `save-markdown`)
  - `journal` (aliases: `get-journaling-by-date`, `journaling`)
- low-level:
  - `block create|update|delete|get`
  - `batch run`

## Env vars

Required:
- `ROAM_API_TOKEN`
- `ROAM_API_GRAPH`

Optional:
- `ROAM_API_BASE_URL` (default: `https://api.roamresearch.com/api/graph`)
- `ROAM_TIMEOUT_SECONDS` (default: `10`)
- `TOPIC_NODE` / `ROAM_TOPIC_NODE` (used by `journal`)

## Build

```bash
cd roamresearch-skills
go mod tidy
go build -o roam-cli ./cmd/roam-cli
```

Or with Makefile:

```bash
cd roamresearch-skills
make build
make run ARGS='--help'
make ci
```

## Test

```bash
cd roamresearch-skills
# unit tests
go test ./...

# bdd tests
go test -tags=bdd ./tests/bdd/...
```

## CI / Release Automation

In this standalone repo, workflows live under `.github/workflows`:
- `go-ci.yml`: format/vet/unit/bdd/build
- `release-on-tag.yml`: auto GitHub Release on tag `roam-cli-v*`
- `release-command.yml`: trigger release tag by PR comment command

PR comment release command:

```text
!release <patch|minor|major> [alpha|beta|rc]
```

Examples:
- `!release patch`
- `!release minor beta`

Manual dispatch is also supported in Actions (`Release Command` workflow).

## Release (local)

```bash
cd roamresearch-skills
./scripts/release.sh v0.1.0
```

This generates cross-platform artifacts in `dist/`:
- darwin amd64/arm64
- linux amd64/arm64
- sha256 checksums

## Examples

```bash
# get by page title or uid
./roam-cli get "Page Title"
./roam-cli get "((block-uid))"

# search
./roam-cli search term1 term2 --limit 20

# datalog query
./roam-cli q '[:find ?title :where [?e :node/title ?title]]'
./roam-cli q "$(cat ./examples/query.page-titles.datalog)"

# save markdown
./roam-cli save --title "New Page" --file ./examples/note.md
cat ./examples/note.md | ./roam-cli save --title "New Page"

# journal
./roam-cli journal --date 2026-02-12

# low-level block create
./roam-cli block create --parent "02-12-2026" --text "hello from go"

# low-level batch from file
./roam-cli batch run --file ./examples/actions.create-page-and-block.json
```
