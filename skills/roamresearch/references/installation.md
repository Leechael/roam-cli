# Installation & Configuration

## Install roam-cli

Preferred: install from GitHub Releases.

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

## Configure credentials

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

## Verify status

```bash
roam-cli status
```

If credentials are missing or invalid, this command will fail with guidance. Do not continue with write operations until `status` is successful.

## Recommended: 1Password CLI for credentials

Prefer running `roam-cli` with the 1Password CLI (`op`) so credentials are injected at runtime instead of stored in shell profiles.

Reference: https://developer.1password.com/docs/service-accounts/use-with-1password-cli

```bash
op run --env-file=.env -- roam-cli status
op run --env-file=.env -- roam-cli get "Page Title"
```

Your `.env` should define `ROAM_API_TOKEN` and `ROAM_API_GRAPH` with 1Password secret references.
