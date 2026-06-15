# Configuration

## Environment Variables

`roam-cli` reads credentials from environment variables:

| Variable | Required | Description |
|---|---|---|
| `ROAM_API_TOKEN` | Yes | Roam Research API token |
| `ROAM_API_GRAPH` | Yes | Roam Research graph name |
| `ROAM_API_BASE_URL` | No | Custom API base URL (default: `https://api.roamresearch.com/api/graph`) |
| `ROAM_TIMEOUT_SECONDS` | No | Request timeout in seconds (default: 30) |

## Secret Management

### Path A: op run

Inject credentials from a `.env` file using 1Password CLI:

```bash
op run --env-file=.env -- roam-cli status
```

### Path B: 1Password Shell Plugin

Prerequisites: [1Password CLI](https://1password.com/downloads/command-line) installed and a 1Password item storing your Roam Research credentials.

Install the local plugin:

```bash
roam-cli onepassword install
op plugin init roam-cli
source ~/.config/op/plugins.sh
roam-cli status
```

> `op plugin init roam-cli` only works after `roam-cli onepassword install` has copied the local plugin binary to `~/.op/plugins/local/roamresearch`.
