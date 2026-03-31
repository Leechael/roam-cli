package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Leechael/roam-cli/internal/client"
	"github.com/Leechael/roam-cli/internal/config"
	"github.com/Leechael/roam-cli/internal/model"
)

type globalOptions struct {
	token   string
	graph   string
	baseURL string
	timeout int
}

var opts globalOptions

// Version is set at build time via -ldflags.
var Version = "dev"

// NewRootCmd creates the root cobra command.
func NewRootCmd() *cobra.Command {
	return newRootCmd()
}

// ExitCode maps an error to a stable exit code.
//
//	0 = success
//	1 = general error
//	2 = auth failure (401/403)
//	3 = not found (404)
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var apiErr *model.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.Status {
		case 401, 403:
			return 2
		case 404:
			return 3
		}
	}
	return 1
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:                   "roam-cli",
		Short:                 "Roam Research CLI (Go)",
		Version:               Version,
		SilenceUsage:          true,
		SilenceErrors:         true,
		CompletionOptions:     cobra.CompletionOptions{HiddenDefaultCmd: true},
	}

	root.PersistentFlags().StringVar(&opts.token, "token", "", "Roam API token (overrides ROAM_API_TOKEN)")
	root.PersistentFlags().StringVar(&opts.graph, "graph", "", "Roam graph name (overrides ROAM_API_GRAPH)")
	root.PersistentFlags().StringVar(&opts.baseURL, "base-url", "", "Roam API base URL (overrides ROAM_API_BASE_URL)")
	root.PersistentFlags().IntVar(&opts.timeout, "timeout", 0, "Request timeout in seconds (overrides ROAM_TIMEOUT_SECONDS)")

	// Hide connection flags from root --help; documented in "roam-cli help configuration"
	root.PersistentFlags().MarkHidden("token")
	root.PersistentFlags().MarkHidden("graph")
	root.PersistentFlags().MarkHidden("base-url")
	root.PersistentFlags().MarkHidden("timeout")

	root.AddGroup(
		&cobra.Group{ID: "daily", Title: "Daily Use:"},
		&cobra.Group{ID: "lowlevel", Title: "Low-level API:"},
	)

	// Daily Use commands
	addToGroup(root, "daily", newSaveCmd())
	addToGroup(root, "daily", newGetCmd())
	addToGroup(root, "daily", newSearchCmd())
	addToGroup(root, "daily", newJournalCmd())
	addToGroup(root, "daily", newMoveCmd())
	addToGroup(root, "daily", newStatusCmd())

	// Low-level API commands
	addToGroup(root, "lowlevel", newBlockCmd())
	addToGroup(root, "lowlevel", newBatchCmd())
	addToGroup(root, "lowlevel", newQCmd())

	installHelpCmd(root)


	return root
}

func addToGroup(parent *cobra.Command, groupID string, child *cobra.Command) {
	child.GroupID = groupID
	parent.AddCommand(child)
}

func mustClient() (*client.Client, error) {
	cfg, err := config.New(opts.token, opts.graph, opts.baseURL, opts.timeout)
	if err != nil {
		return nil, err
	}
	return client.NewClient(cfg), nil
}

func errMissingFlag(name string) error {
	return fmt.Errorf("missing required flag: --%s", name)
}
