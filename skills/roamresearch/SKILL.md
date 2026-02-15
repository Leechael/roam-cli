---
name: roamresearch
description: Roam Research operations via roam-cli. Use this skill for page/block retrieval, search, datalog queries, markdown save, journaling lookup, and low-level block or batch writes with environment-injected credentials.
compatibility: Requires roam-cli in PATH and ROAM_API_TOKEN/ROAM_API_GRAPH environment variables.
---

# RoamResearch Skill

Use this skill to perform Roam Research read/write/query workflows through `roam-cli`.

## Prerequisites

### 1) Ensure `roam-cli` is installed

Preferred: install from GitHub Releases of this repository.

```bash
# 1) Inspect releases
gh release list -R Leechael/roamresearch-skills

# 2) Download latest artifacts
gh release download -R Leechael/roamresearch-skills --pattern 'roam-cli-*.tar.gz'

# 3) Extract your platform artifact and install binary
tar -xzf roam-cli-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/').tar.gz
install -m 0755 roam-cli /usr/local/bin/roam-cli
```

Quick check:

```bash
roam-cli --help
```

### 2) Configure credentials

Required environment variables:

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

### 3) Verify login/config status before operations

Use the built-in status check command:

```bash
roam-cli status
```

If credentials are missing or invalid, this command will fail with guidance. Do not continue with write operations until `status` is successful.

## Recommended Credential Management (1Password CLI)

Prefer running `roam-cli` with the 1Password CLI (`op`) so credentials are injected at runtime instead of stored in shell profiles.

Reference:
- https://developer.1password.com/docs/service-accounts/use-with-1password-cli

Example pattern:

```bash
op run --env-file=.env -- roam-cli status
op run --env-file=.env -- roam-cli get "Page Title"
```

(Your `.env` should define `ROAM_API_TOKEN` and `ROAM_API_GRAPH` with 1Password secret references.)

## Command Mapping

- Get page/block: `roam-cli get`
- Search blocks: `roam-cli search`
- Run datalog query: `roam-cli q`
- Save markdown page: `roam-cli save` (alias: `save-markdown`)
- Get journaling by date: `roam-cli journal`
- Find block UID: `roam-cli block find`
- Create nested block tree: `roam-cli block create-tree`
- Low-level block API: `roam-cli block ...`
- Low-level batch API: `roam-cli batch run ...`

## Recommended Workflow

1. Run `roam-cli status` first.
2. Prefer read commands first: `get`, `search`, `q`.
3. Prefer high-level writes: `save`, `journal`.
4. Use low-level APIs for deterministic control: `block`, `batch`.

## Usage Examples

### 1) Read page/block

```bash
roam-cli get "Page Title"
roam-cli get "((block-uid))"
roam-cli get "Page Title" --raw
```

### 2) Search blocks

```bash
roam-cli search term1 term2 --limit 20
roam-cli search keyword --page "Project" --ignore-case
```

### 3) Datalog query

```bash
roam-cli q '[:find ?title :where [?e :node/title ?title]]'
```

### 4) Save markdown

```bash
roam-cli save --title "New Page" --file ./note.md
cat ./note.md | roam-cli save --title "New Page"
```

### 5) Journal by date

```bash
roam-cli journal --date 2026-02-12
roam-cli journal --date 2026-02-12 --topic "Work Log"
```

### 6) Find block UID

```bash
# Find block by text on a daily note
roam-cli block find --text "[[📖 Daily Reading]]" --daily 2026-02-15

# Find block by text on a named page
roam-cli block find --text "Status" --page "Project Dashboard"
```

### 7) Create nested block tree

```bash
# From stdin (single object)
echo '{"text":"headline","children":[{"text":"snapshot"}]}' | roam-cli block create-tree --parent <uid> --stdin

# From stdin (array)
echo '[{"text":"item1"},{"text":"item2","children":[{"text":"sub"}]}]' | roam-cli block create-tree --parent <uid> --stdin

# From file
roam-cli block create-tree --parent <uid> --file ./tree.json
```

### 8) Optimized daily-note workflow (2 calls)

```bash
# Step 1: Find the target block UID
UID=$(roam-cli block find --daily 2026-02-15 --text "[[📖 Daily Reading]]")

# Step 2: Create nested tree under it
echo '{"text":"headline","children":[{"text":"snapshot"}]}' | roam-cli block create-tree --parent "$UID" --stdin
```

### 9) Low-level block operations

```bash
roam-cli block create --parent <uid> --text "hello"
roam-cli block update --uid <uid> --text "updated"
roam-cli block delete --uid <uid>
roam-cli block get --uid <uid>
```

### 10) Low-level batch operations

```bash
roam-cli batch run --file ./examples/actions.create-page-and-block.json
cat ./examples/actions.create-page-and-block.json | roam-cli batch run --stdin
```

## Error Handling Rules

- Missing credentials: explicitly report missing `ROAM_API_TOKEN` / `ROAM_API_GRAPH`.
- API failures: include HTTP status code and response body.
- Not found: clearly include the identifier/uid that was requested.

## Output Rules

- Preserve JSON output when `--raw` is requested.
- Keep default output concise and readable.
- Never invent Roam data; only report real command results.
