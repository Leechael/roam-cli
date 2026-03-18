package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"roam-cli/internal/config"
	"roam-cli/internal/roam"
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

func Execute() error {
	return newRootCmd().Execute()
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:     "roam-cli",
		Short:   "Roam Research CLI (Go)",
		Version: Version,
	}

	root.PersistentFlags().StringVar(&opts.token, "token", "", "Roam API token (overrides ROAM_API_TOKEN)")
	root.PersistentFlags().StringVar(&opts.graph, "graph", "", "Roam graph name (overrides ROAM_API_GRAPH)")
	root.PersistentFlags().StringVar(&opts.baseURL, "base-url", "", "Roam API base URL (overrides ROAM_API_BASE_URL)")
	root.PersistentFlags().IntVar(&opts.timeout, "timeout", 0, "Request timeout in seconds (overrides ROAM_TIMEOUT_SECONDS)")

	root.AddCommand(newStatusCmd())
	root.AddCommand(newGetCmd())
	root.AddCommand(newSearchCmd())
	root.AddCommand(newSearchPagesCmd())
	root.AddCommand(newQCmd())
	root.AddCommand(newSaveCmd())
	root.AddCommand(newJournalCmd())
	root.AddCommand(newBlockCmd())
	root.AddCommand(newBatchCmd())

	installHelpCmd(root)

	return root
}

func mustClient() (*roam.Client, error) {
	cfg, err := config.New(opts.token, opts.graph, opts.baseURL, opts.timeout)
	if err != nil {
		return nil, err
	}
	return roam.NewClient(cfg), nil
}

func errMissingFlag(name string) error {
	return fmt.Errorf("missing required flag: --%s", name)
}
