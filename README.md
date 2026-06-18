# roam-cli

`roam-cli` is a command-line tool for working with your Roam Research graph.

- **Daily Use**: Read pages, search content, save markdown, and manage daily notes
- **Low-level API**: Direct access to blocks, batch actions, and Datalog queries

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

## Quick Start

```bash
# Verify setup
roam-cli status

# Read today's daily page
roam-cli get --today

# Quick capture to journal
printf '- Had a great idea' | roam-cli save --today --under '[[📽 Journaling]]'

# Search your graph
roam-cli search "meeting" "action item"
```

---

## Configuration

Set credentials via environment variables (recommended):

```bash
export ROAM_API_TOKEN="<token>"
export ROAM_API_GRAPH="<graph>"
```

Or use flags for one-off commands:

```bash
roam-cli --graph other-graph --token <token> get "Page Title"
```

### Optional settings

| Environment Variable | Flag | Default | Description |
|---|---|---|---|
| `ROAM_API_BASE_URL` | `--base-url` | `https://api.roamresearch.com/api/graph` | API endpoint |
| `ROAM_TIMEOUT_SECONDS` | `--timeout` | `10` | Request timeout |

### Get your API token

1. Open Roam Research → Settings → Graph → API tokens
2. Click "New API Token", give it a name, select read+write access
3. Copy the token value

### Recommended: 1Password CLI

```bash
# .env file
ROAM_API_TOKEN=op://vault/roam-api/token
ROAM_API_GRAPH=op://vault/roam-api/graph

# Run with credential injection
op run --env-file=.env -- roam-cli status
```

---

## Commands

### Daily Use Commands

| Command | Purpose |
|---------|---------|
| `status` | Check credentials and API connectivity |
| `get` | Read page by title, block by uid, or daily page by date |
| `search` | Search blocks or pages by terms |
| `save` | Save GFM markdown as a Roam page or daily page section |
| `page` | Clear or delete a page by title or date |
| `journal` | Get journaling blocks from Daily Notes |
| `move` | Move a block to a page or section |
| `help` | Show help for commands or topics |

### Low-level API Commands

| Command | Purpose |
|---------|---------|
| `q` | Execute raw Datalog query |
| `block` | Low-level block APIs (create, update, delete, move, get, find) |
| `batch` | Low-level batch actions API |

### Help Topics

```bash
roam-cli help configuration     # Setup and credentials
roam-cli help format           # GFM to Roam conversion rules
roam-cli help writing-guide    # Choose the right write command
roam-cli help datalog          # Datalog query reference
roam-cli help exit-codes       # Exit codes for scripting
roam-cli help read-examples    # Reading examples
roam-cli help write-examples   # Writing examples
roam-cli help workflow-examples # Common workflows
```

---

## Output Modes

- `--json` — Parseable JSON output
- `--plain` — Human-readable plain text
- `--jq '<expr>'` — Filter JSON with jq expression (requires `--json`)

`--json` and `--plain` are mutually exclusive.

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Authentication failure (401/403) |
| 3 | Resource not found (404) |

---

## Usage Examples

### Reading

```bash
# Today's daily page
roam-cli get --today
roam-cli get --today --json

# Daily page by date
roam-cli get --daily yesterday
roam-cli get --daily 2026-03-14

# Page by title
roam-cli get "Page Title"

# Block by uid
roam-cli get "((block-uid))"

# Journaling blocks
roam-cli journal --date today
roam-cli journal --date yesterday --topic "Work Log"
```

### Searching

```bash
# Search pages (default)
roam-cli search "meeting" "action item" --limit 10

# Search individual blocks
roam-cli search keyword --type block --limit 20

# Search in specific page or daily page
roam-cli search keyword --page "Project"
roam-cli search "TODO" --daily-topic "[[TODO]]"
```

### Saving Markdown

```bash
# Save to today's daily page
printf '- journal entry' | roam-cli save --today

# Save under a section (find-or-create)
printf '- entry' | roam-cli save --today --under '[[📽 Journaling]]'

# Save to a named page
cat note.md | roam-cli save --title "New Page"

# Replace a named page: clear existing page blocks first, then write note.md.
# Without --replace, save appends new blocks to the page.
cat note.md | roam-cli save --title "Project X" --replace

# Save under a section in a named page
printf '- new task' | roam-cli save --title "Project X" --under '[[TODO]]'

# Save to a specific daily page
cat note.md | roam-cli save --to-daily-page 2026-03-14

# Save under existing parent block (by UID)
roam-cli save --parent <uid> --file ./note.md

# Page operations
roam-cli page clear "Project X"
roam-cli page delete "Project X"

# Get UID back for follow-up commands
UID=$(printf '- item' | roam-cli save --today --under '[[Inbox]]' --plain)
printf '- detail' | roam-cli save --parent "$UID"
```

### Moving Blocks

```bash
# Move block to a page section
roam-cli move --uid <block> --title "Project X" --under "[[Tasks]]"

# Move to today's daily page section
roam-cli move --uid <block> --today --under "[[Archive]]"

# Move to daily page by date
roam-cli move --uid <block> --daily yesterday
```

### Datalog Queries

```bash
# List all page titles
roam-cli q '[:find ?title :where [?e :node/title ?title]]'

# Find blocks containing text (case-insensitive)
roam-cli q '[:find ?uid ?s
 :where
   [?b :block/string ?s]
   [?b :block/uid ?uid]
   [(clojure.string/includes? (clojure.string/lower-case ?s) "meeting")]]' --json

# Find blocks matching regex
roam-cli q '[:find (pull ?node [:block/uid :block/string])
 :where
   [(re-pattern "https?://x\\.com") ?pattern]
   [?node :block/string ?text]
   [(re-find ?pattern ?text)]]' --json

# Pull full page tree with children
roam-cli q '[:find (pull ?e [* {:block/children ...}])
 :where [?e :node/title "My Page"]]' --json
```

### Low-level Block Operations

```bash
# Create single block
roam-cli block create --parent <uid> --text "hello"

# Create nested tree from JSON
echo '{"text":"Root","children":[{"text":"Child"}]}' \
  | roam-cli block create --parent <uid>

# Create under a section (find-or-create)
roam-cli block create --parent <page-uid> --attach-to "[[Section]]" --text "item"

# Find block by text
roam-cli block find --text "[[📖 Daily Reading]]" --today

# Update block
roam-cli block update --uid <uid> --text "updated"

# Move block
roam-cli block move --uid <uid> --parent <target-uid> --order last

# Delete block
roam-cli block delete --uid <uid>
```

### Batch Operations

```bash
# Run batch from file
roam-cli batch run --file ./actions.json

# Run batch from stdin
echo '[...]' | roam-cli batch run
```

---

## GFM Format Reference

The `save` command converts GitHub Flavored Markdown to Roam blocks:

| GFM element | Roam result |
|---|---|
| `#` (h1) | Skipped — page title comes from `--title` |
| `##` – `###` | Block with heading attribute |
| `- item` / `* item` | Nested child block |
| Indented list items | Deeper child blocks |
| `1. item` | Child block with marker preserved |
| ` ```lang ... ``` ` | Single block with fenced code |
| `> quote` | Block with `> ` prefix |
| `---` / `***` | Discarded |
| GFM table | `{{[[table]]}}` with row/cell children |
| Consecutive text lines | Joined into one paragraph block |

Inline formatting (bold, italic, code, links) passes through as-is.

---

## Install the Agent Skill

This repository ships an Agent Skill package under `skills/roamresearch`.

Install with:

```bash
npx skills add Leechael/roamresearch-skills
```

After installation, your agent can load and use the `roamresearch` skill instructions.

---

## Tips

- Use `printf` instead of `--text` for content with `[[ ]]` or emoji — it preserves special characters
- `save --plain` outputs the first saved block UID, useful for composing commands
- Build trees in one call with JSON input instead of sequential `block create` calls
- Use `roam-cli help <topic>` for detailed documentation on specific topics

### 1Password shell plugin

The shell plugin lets 1Password CLI automatically inject credentials when you run `roam-cli`.

Prerequisites:
- [1Password CLI](https://1password.com/downloads/command-line) installed
- A 1Password item with fields matching the environment variables below

Install the local plugin:

```bash
# Install the 1Password shell plugin binary.
roam-cli onepassword install

# Let 1Password create the shell wrapper.
op plugin init roam-cli

# Reload shell plugin aliases.
source ~/.config/op/plugins.sh

# Run normally.
roam-cli status
```

> `op plugin init roam-cli` only works after `roam-cli onepassword install` has copied the local plugin binary to `~/.op/plugins/local/roamresearch`.

#### Release archives

Starting from the release that includes this feature, each release archive contains both `roam-cli` and `roamresearch` binaries. Extract both to your `PATH` before running `roam-cli onepassword install`.

#### Developer flow

```bash
make build op-plugin-build
./bin/roam-cli onepassword install --from ./bin/roamresearch --force
```

#### Manual shell wrapper

If you prefer not to use `op plugin init`, add this to your shell rc file:

```bash
roam-cli() {
  op plugin run -- roam-cli "$@"
}
export OP_PLUGIN_ALIASES_SOURCED=1
```
