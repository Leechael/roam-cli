package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// execLookPath is overridable in tests.
var execLookPath = exec.LookPath

const (
	localPluginDir    = "/.op/plugins/local"
	localPluginPath   = "~/.op/plugins/local/roamresearch"
	localPluginBinary = "roamresearch"
)

type onePasswordOptions struct {
	from  string
	force bool
}

var opOpts onePasswordOptions

// defaultOPPluginDir returns the default local plugin directory.
// Override in tests by setting testOPPluginDir.
var testOPPluginDir string

func opPluginDir() string {
	if testOPPluginDir != "" {
		return testOPPluginDir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home + localPluginDir
}

func newOnePasswordCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "onepassword",
		Short: "Manage 1Password shell plugin integration",
	}

	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install the 1Password shell plugin for roam-cli",
		Long: `Install the 1Password shell plugin binary to the local plugins directory.

This copies the roamresearch plugin binary to ~/.op/plugins/local/roamresearch
so that 1Password CLI can discover it.

After installation, run:
  op plugin list | grep roam-cli
  op plugin init roam-cli
  source ~/.config/op/plugins.sh`,
		RunE: runOnePasswordInstall,
	}

	installCmd.Flags().StringVar(&opOpts.from, "from", "", "Copy plugin binary from PATH instead of auto-discovery")
	installCmd.Flags().BoolVar(&opOpts.force, "force", false, "Replace existing ~/.op/plugins/local/roamresearch")

	cmd.AddCommand(installCmd)
	return cmd
}

func findOpBinary() (string, error) {
	path, err := execLookPath("op")
	if err != nil {
		return "", fmt.Errorf("1Password CLI (op) not found in PATH; install it from https://1password.com/downloads/command-line")
	}
	return path, nil
}

func resolvePluginSource(fromFlag string) (string, error) {
	if fromFlag != "" {
		info, err := os.Stat(fromFlag)
		if err != nil {
			return "", fmt.Errorf("--from path %q does not exist", fromFlag)
		}
		if info.Mode()&0o100 == 0 {
			return "", fmt.Errorf("--from path %q is not executable", fromFlag)
		}
		return fromFlag, nil
	}

	// Look next to the running binary.
	exe, err := os.Executable()
	if err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "roamresearch")
		if info, err := os.Stat(candidate); err == nil && info.Mode().IsRegular() && info.Mode()&0o100 != 0 {
			return candidate, nil
		}
	}

	// Look in PATH.
	candidate, err := execLookPath("roamresearch")
	if err == nil {
		return candidate, nil
	}

	return "", fmt.Errorf("roamresearch plugin binary not found; place it next to roam-cli or in PATH, or use --from")
}

func runOnePasswordInstall(cmd *cobra.Command, args []string) error {
	if _, err := findOpBinary(); err != nil {
		return err
	}

	src, err := resolvePluginSource(opOpts.from)
	if err != nil {
		return err
	}

	destDir := opPluginDir()
	if destDir == "" {
		return fmt.Errorf("cannot determine home directory")
	}
	dest := filepath.Join(destDir, localPluginBinary)

	if err := ensureLocalPluginDirs(destDir); err != nil {
		return err
	}

	if info, err := os.Stat(dest); err == nil && !opOpts.force {
		if info.Mode().IsRegular() {
			return fmt.Errorf("%s already exists; use --force to overwrite", dest)
		}
	}

	if err := copyFile(src, dest); err != nil {
		return fmt.Errorf("cannot copy plugin binary: %w", err)
	}

	if err := os.Chmod(dest, 0755); err != nil {
		return fmt.Errorf("cannot set permissions on %s: %w", dest, err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Installed 1Password shell plugin: %s\n", localPluginPath)
	fmt.Fprintln(cmd.OutOrStdout(), "Next steps:")
	fmt.Fprintln(cmd.OutOrStdout(), "  op plugin list | grep roam-cli")
	fmt.Fprintln(cmd.OutOrStdout(), "  op plugin init roam-cli")
	fmt.Fprintln(cmd.OutOrStdout(), "  source ~/.config/op/plugins.sh")

	return nil
}

func ensureLocalPluginDirs(destDir string) error {
	if err := os.MkdirAll(destDir, 0700); err != nil {
		return fmt.Errorf("cannot create plugin directory %s: %w", destDir, err)
	}

	for _, dir := range []string{filepath.Dir(filepath.Dir(destDir)), filepath.Dir(destDir), destDir} {
		if err := os.Chmod(dir, 0700); err != nil {
			return fmt.Errorf("cannot set permissions on %s: %w", dir, err)
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

// RegisterOnepasswordCmd adds the onepassword subcommand to the given root command.
func RegisterOnepasswordCmd(root *cobra.Command) {
	root.AddCommand(newOnePasswordCmd())
}
