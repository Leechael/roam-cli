# Configuration

How to set up credentials and connection options for roam-cli.

## Overview

roam-cli needs two pieces of information to connect to your Roam graph:
an API token and a graph name. These can be provided via environment variables
or command-line flags. Environment variables are recommended for daily use;
flags are useful for one-off overrides.

## Required

| Environment Variable | Flag | Description |
|---|---|---|
| `ROAM_API_TOKEN` | `--token` | Roam Research API token |
| `ROAM_API_GRAPH` | `--graph` | Roam graph name |

## Optional

| Environment Variable | Flag | Default | Description |
|---|---|---|---|
| `ROAM_API_BASE_URL` | `--base-url` | `https://api.roamresearch.com/api/graph` | API endpoint |
| `ROAM_TIMEOUT_SECONDS` | `--timeout` | `10` | Request timeout in seconds |

## Commands

- `roam-cli status` -- verify credentials and API connectivity

## Setup

### Quick start

```bash
export ROAM_API_TOKEN="roam-graph-token-..."
export ROAM_API_GRAPH="your-graph-name"
roam-cli status
```

### Get your API token

1. Open Roam Research in your browser
2. Go to Settings (top-right menu) > Graph > API tokens
3. Click "New API Token", give it a name, select read+write access
4. Copy the token value

### Recommended: 1Password CLI

Store credentials in 1Password and inject at runtime instead of keeping
them in shell profiles:

```bash
# Create a .env file with 1Password secret references
cat > .env <<'EOF'
ROAM_API_TOKEN=op://vault/roam-api/token
ROAM_API_GRAPH=op://vault/roam-api/graph
EOF

# Run with credential injection
op run --env-file=.env -- roam-cli status
op run --env-file=.env -- roam-cli get --today
```

Reference: https://developer.1password.com/docs/service-accounts/use-with-1password-cli

### Shell profile

For personal machines without a secrets manager:

```bash
# Add to ~/.bashrc or ~/.zshrc
export ROAM_API_TOKEN="roam-graph-token-..."
export ROAM_API_GRAPH="your-graph-name"
```

## Constraints

- API tokens are graph-specific. You need a separate token for each graph.
- Tokens can be revoked from the Roam Settings page at any time.
- The `--token` and `--graph` flags override environment variables when both
  are set. This is useful for scripts that access multiple graphs.
- `roam-cli status` is the canonical way to verify credentials. Run it before
  any other operations.

## Examples

Override graph for a one-off command:

```bash
roam-cli --graph other-graph get "Page Title"
```

Use a longer timeout for large batch operations:

```bash
roam-cli --timeout 60 batch run --file large-import.json
```

## Related Topics

- `roam-cli help writing-guide`
- `roam-cli help exit-codes`
