# 1Password Shell Plugin Support Plan

## Purpose

Add local 1Password Shell Plugin support for `roam-cli` inside this repository. Do not submit this as an official plugin to `1Password/shell-plugins`.

The target user experience is:

```bash
roam-cli onepassword install
op plugin init roam-cli
source ~/.config/op/plugins.sh
roam-cli status
```

## Fixed Decisions

These decisions are part of the implementation contract. Do not rename or substitute them during implementation.

| Topic | Decision |
|---|---|
| Main CLI command | `roam-cli onepassword install` |
| Plugin name / local plugin binary | `roamresearch` |
| Plugin executable usage | `roam-cli` |
| Local install destination | `~/.op/plugins/local/roamresearch` |
| Runtime install strategy | Copy a trusted `roamresearch` binary from the release bundle, same directory as `roam-cli`, `PATH`, or explicit `--from` |
| Network download in installer | No |
| Official 1Password plugin submission | No |
| Current `op run --env-file` support | Keep |

Use `onepassword`, not `1password`, as the CLI command name. Cobra commands that start with a digit are awkward for help, tests, and shell usage.

## Current State

`roam-cli` already reads credentials from environment variables in `internal/config/env.go`:

- `ROAM_API_TOKEN`
- `ROAM_API_GRAPH`
- optional: `ROAM_API_BASE_URL`
- optional: `ROAM_TIMEOUT_SECONDS`

No auth logic change is required in the main client. The shell plugin only injects these environment variables before executing the real `roam-cli` binary.

## 1Password Discovery Model

`op plugin init roam-cli` does not scan this repository or the current working directory.

1Password CLI discovers shell plugins from:

1. plugins bundled with the installed `op` binary
2. local plugin binaries under `~/.op/plugins/local/`

For local plugins, `op` starts the plugin binary and reads its schema over RPC. A source-only plugin in this repository is not enough. The compiled `roamresearch` plugin binary must exist at:

```text
~/.op/plugins/local/roamresearch
```

Only after that should this work:

```bash
op plugin list | grep roam-cli
op plugin init roam-cli
```

## Shell Wrapper Behavior

After initialization, 1Password writes a shell wrapper similar to:

```bash
roam-cli() {
  op plugin run -- roam-cli "$@"
}
```

This shadows the real `roam-cli` binary in the interactive shell. It should not recurse because `op plugin run` is a separate process and resolves the real executable from `PATH`, not from the shell function.

Users can bypass the wrapper by calling the binary with an absolute path:

```bash
/usr/local/bin/roam-cli status
```

## User Flows

### Release user

The release archive must contain both binaries:

```text
roam-cli
roamresearch
```

User flow:

```bash
# Install both binaries by the normal release install path.
roam-cli --version

# Copy the local 1Password plugin to ~/.op/plugins/local/roamresearch.
roam-cli onepassword install

# Confirm op can discover it.
op plugin list | grep roam-cli

# Let 1Password create the shell wrapper.
op plugin init roam-cli

# Reload shell plugin aliases.
source ~/.config/op/plugins.sh

# Run normally.
roam-cli status
```

### Source checkout / developer

```bash
make op-plugin-build
roam-cli onepassword install --from ./bin/roamresearch
op plugin list | grep roam-cli
op plugin init roam-cli
```

### Dotfile-managed shell wrapper

Users who do not want `op plugin init` to edit `~/.config/op/plugins.sh` can add this manually:

```bash
roam-cli() {
  op plugin run -- roam-cli "$@"
}

export OP_PLUGIN_ALIASES_SOURCED=1
```

## Repository Changes

Implement the plugin as a nested Go module:

```text
contrib/1password-plugin/
  go.mod
  go.sum
  main.go
  plugin.go
  credential.go
  executable.go
  credential_test.go
  plugin_test.go
```

Update the main CLI:

```text
internal/cmd/onepassword.go
internal/cmd/root.go
internal/cmd/onepassword_test.go
```

Update build/docs:

```text
Makefile
README.md
docs/help/topics/configuration.md
```

Do not add a runtime dependency on repository-local scripts. Installed users may only have binaries, not the source checkout.

## Plugin Module Setup

Create the nested module with:

```bash
mkdir -p contrib/1password-plugin
cd contrib/1password-plugin
go mod init github.com/Leechael/roam-cli/contrib/1password-plugin
go get github.com/1Password/shell-plugins@af9327a
go mod tidy
```

The exact pseudo-version written to `go.mod` may differ after `go get`. Keep the generated `go.mod` and `go.sum`.

## Plugin Schema

Credential type: `API Token`

Use these field names exactly:

```go
const (
	fieldToken          = sdk.FieldName("Token")
	fieldGraph          = sdk.FieldName("Graph")
	fieldAPIURL         = sdk.FieldName("API URL")
	fieldTimeoutSeconds = sdk.FieldName("Timeout Seconds")
)
```

Do not use non-existent constants such as `fieldname.Graph` or `fieldname.TimeoutSeconds`.

Environment mapping:

```go
var envVarMapping = map[string]sdk.FieldName{
	"ROAM_API_TOKEN":       fieldToken,
	"ROAM_API_GRAPH":       fieldGraph,
	"ROAM_API_BASE_URL":    fieldAPIURL,
	"ROAM_TIMEOUT_SECONDS": fieldTimeoutSeconds,
}
```

Fields:

| Field | Env var | Secret | Optional |
|---|---|---:|---:|
| Token | `ROAM_API_TOKEN` | yes | no |
| Graph | `ROAM_API_GRAPH` | no | no |
| API URL | `ROAM_API_BASE_URL` | no | yes |
| Timeout Seconds | `ROAM_TIMEOUT_SECONDS` | no | yes |

Credential implementation rules:

- `Token` is the only secret field.
- `Graph` is required because `roam-cli` fails without it.
- `API URL` and `Timeout Seconds` are optional.
- Use `provision.EnvVars(envVarMapping)` as `DefaultProvisioner`.
- Use `importer.TryEnvVarPair(envVarMapping)` as `Importer`.

Executable schema:

```go
schema.Executable{
	Name:    "Roam Research CLI",
	Runs:    []string{"roam-cli"},
	DocsURL: sdk.URL("https://github.com/Leechael/roam-cli"),
	NeedsAuth: needsauth.IfAll(
		needsauth.NotForHelpOrVersion(),
		needsauth.NotWithoutArgs(),
	),
	Uses: []schema.CredentialUsage{
		{Name: credname.APIToken},
	},
}
```

Plugin schema:

```go
schema.Plugin{
	Name: "roamresearch",
	Platform: schema.PlatformInfo{
		Name:     "Roam Research",
		Homepage: sdk.URL("https://roamresearch.com"),
	},
	Credentials: []schema.CredentialType{APIToken()},
	Executables: []schema.Executable{RoamCLI()},
}
```

## Plugin Binary RPC Server

`contrib/1password-plugin/main.go` must start a 1Password shell plugin RPC server. A schema-only `main.go` is not enough.

Use this structure:

```go
package main

import (
	"github.com/1Password/shell-plugins/sdk/rpc/proto"
	"github.com/1Password/shell-plugins/sdk/rpc/server"
	"github.com/1Password/shell-plugins/sdk/schema"
	"github.com/hashicorp/go-plugin"
)

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  proto.Version,
			MagicCookieKey:   proto.MagicCookieKey,
			MagicCookieValue: proto.MagicCookieValue,
		},
		Plugins: plugin.PluginSet{
			"plugin": &server.RPCPlugin{RPCPlugin: func() (schema.Plugin, error) {
				return New(), nil
			}},
		},
	})
}
```

## Main CLI Install Command

Add `roam-cli onepassword install`.

Flags:

```text
--from PATH   Copy plugin binary from PATH instead of auto-discovery
--force       Replace existing ~/.op/plugins/local/roamresearch
```

Behavior:

1. If `op` is not in `PATH`, return an error that asks the user to install 1Password CLI.
2. Resolve the source plugin binary:
   1. if `--from` is set, use that path
   2. otherwise look for `roamresearch` in the same directory as the running `roam-cli` binary
   3. otherwise look for `roamresearch` in `PATH`
   4. otherwise fail with a message that the release must include `roamresearch`
3. Verify the source path exists and is executable.
4. Create these directories if missing:
   - `~/.op`
   - `~/.op/plugins`
   - `~/.op/plugins/local`
5. Set directory permissions to `0700` where possible.
6. If `~/.op/plugins/local/roamresearch` exists and `--force` is not set, return an error explaining `--force`.
7. Copy the binary to `~/.op/plugins/local/roamresearch`.
8. Set destination mode to `0755`.
9. Print exactly these next steps:

```text
Installed 1Password shell plugin: ~/.op/plugins/local/roamresearch
Next steps:
  op plugin list | grep roam-cli
  op plugin init roam-cli
  source ~/.config/op/plugins.sh
```

Do not run `op plugin init` automatically. It mutates user shell configuration.

Do not print secret values or inspect user 1Password items.

## Makefile Changes

Add these variables:

```make
OP_PLUGIN_NAME := roamresearch
OP_PLUGIN_DIR := contrib/1password-plugin
OP_PLUGIN_BIN := $(BIN_DIR)/$(OP_PLUGIN_NAME)
```

Add these targets:

```make
op-plugin-test:
	cd $(OP_PLUGIN_DIR) && go test ./... -count=1

op-plugin-build:
	mkdir -p $(BIN_DIR)
	cd $(OP_PLUGIN_DIR) && go build -o ../../$(OP_PLUGIN_BIN) .

op-plugin-install-local: op-plugin-build
	mkdir -p ~/.op/plugins/local
	chmod 700 ~/.op ~/.op/plugins ~/.op/plugins/local
	cp $(OP_PLUGIN_BIN) ~/.op/plugins/local/$(OP_PLUGIN_NAME)
	chmod 755 ~/.op/plugins/local/$(OP_PLUGIN_NAME)
```

Update `ci` to include `op-plugin-test`.

Update `clean` to remove `$(OP_PLUGIN_BIN)` through the existing `rm -rf $(BIN_DIR)`.

## Release Packaging

Change release archives so each OS/arch archive contains both binaries:

```text
roam-cli
roamresearch
```

Artifact naming:

```text
roam-cli_<os>_<arch>.tar.gz
```

For each target OS/arch:

1. build main binary as `roam-cli`
2. build plugin binary as `roamresearch`
3. tar both files into one archive

Do not publish a standalone plugin-only archive as the primary user path. The install command depends on finding `roamresearch` next to `roam-cli` after normal installation.

If the release workflow cannot bundle both binaries in the first implementation, block and ask before choosing a different distribution path.

## Tests

### Plugin tests

In `contrib/1password-plugin`:

- provisioner maps fields to the four expected env vars
- importer reads the four expected env vars
- needs-auth returns false for:
  - no args
  - `--help`
  - `help`
  - `--version`
- needs-auth returns true for:
  - `status`
  - `get --today`
- schema deep validation has no errors

Run:

```bash
cd contrib/1password-plugin && go test ./... -count=1
```

### Main CLI tests

In `internal/cmd/onepassword_test.go`:

- missing source binary returns a clear error
- existing destination without `--force` returns a clear error
- `--from` copies to a temp fake OP config directory
- installed file mode is executable

Implementation note: do not write tests to the real `~/.op`. Add a small helper that accepts an overridable config dir, or use an unexported package variable reset with `t.Cleanup`.

### Full local manual test

Only run this on a machine with 1Password CLI installed:

```bash
make build op-plugin-build
./bin/roam-cli onepassword install --from ./bin/roamresearch --force
op plugin list | grep roam-cli
op plugin init roam-cli
op plugin run -- roam-cli --help
```

Real credential test is manual because it requires a valid 1Password item and Roam token:

```bash
op plugin run -- roam-cli status
```

Do not put the real credential test in CI.

## Documentation Updates

Update `README.md` and `docs/help/topics/configuration.md` with two supported 1Password paths.

Path A: simple existing `op run` usage:

```bash
op run --env-file=.env -- roam-cli status
```

Path B: shell plugin usage:

```bash
roam-cli onepassword install
op plugin init roam-cli
source ~/.config/op/plugins.sh
roam-cli status
```

Explain that `op plugin init roam-cli` only works after `roam-cli onepassword install` has copied the local plugin binary.

## Implementation Order

Follow this order exactly:

1. Add `contrib/1password-plugin` module and plugin schema.
2. Add plugin tests and make `cd contrib/1password-plugin && go test ./...` pass.
3. Add `Makefile` targets: `op-plugin-test`, `op-plugin-build`, `op-plugin-install-local`.
4. Add `roam-cli onepassword install` command and tests.
5. Update root `ci` target to include plugin tests.
6. Update docs.
7. Update release packaging.
8. Run final validation.

Do not change the existing Roam API client auth behavior unless a test proves it is required.

## Final Validation

Before reporting completion, run:

```bash
go test ./... -count=1
cd contrib/1password-plugin && go test ./... -count=1
make op-plugin-build
prek
```

If `op` is installed locally, also run:

```bash
./bin/roam-cli onepassword install --from ./bin/roamresearch --force
op plugin list | grep roam-cli
op plugin run -- roam-cli --help
```
