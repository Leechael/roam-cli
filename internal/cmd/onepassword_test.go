package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestPluginDir creates a temp directory structure mimicking ~/.op/plugins/local.
func setupTestPluginDir(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	pluginDir := filepath.Join(tmpDir, ".op", "plugins", "local")
	err := os.MkdirAll(pluginDir, 0700)
	if err != nil {
		t.Fatal(err)
	}
	old := testOPPluginDir
	testOPPluginDir = pluginDir
	return tmpDir, func() { testOPPluginDir = old }
}

// writeTempPluginBinary writes an executable file for use as a source plugin binary.
func writeTempPluginBinary(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	content := []byte("#!/bin/sh\necho mock-plugin\n")
	if err := os.WriteFile(path, content, 0755); err != nil {
		t.Fatal(err)
	}
	return path
}

// mockOpLookup replaces execLookPath so that "op" is always found.
func mockOpLookup() func() {
	orig := execLookPath
	execLookPath = func(name string) (string, error) {
		if name == "op" {
			return "/usr/local/bin/op", nil
		}
		return "", os.ErrNotExist
	}
	return func() { execLookPath = orig }
}

func TestOnePasswordInstall_MissingSourceBinary(t *testing.T) {
	_, cleanup := setupTestPluginDir(t)
	defer cleanup()
	defer mockOpLookup()()

	cmd := newOnePasswordCmd()
	installCmd := cmd.Commands()[0]
	opOpts.from = ""
	opOpts.force = false

	err := installCmd.RunE(installCmd, nil)
	if err == nil {
		t.Fatal("expected error for missing source binary")
	}
	if !strings.Contains(err.Error(), "roamresearch plugin binary not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestOnePasswordInstall_ExistingDestinationWithoutForce(t *testing.T) {
	_, cleanup := setupTestPluginDir(t)
	defer cleanup()
	defer mockOpLookup()()

	// Create an existing plugin in the destination.
	destPath := filepath.Join(testOPPluginDir, "roamresearch")
	if err := os.WriteFile(destPath, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a source binary.
	srcDir := t.TempDir()
	srcPath := writeTempPluginBinary(t, srcDir, "roamresearch")

	installCmd := newOnePasswordCmd().Commands()[0]
	opOpts.from = srcPath
	opOpts.force = false

	err := installCmd.RunE(installCmd, nil)
	if err == nil {
		t.Fatal("expected error for existing destination without --force")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestOnePasswordInstall_WithFromFlag(t *testing.T) {
	_, cleanup := setupTestPluginDir(t)
	defer cleanup()
	defer mockOpLookup()()

	srcDir := t.TempDir()
	srcPath := writeTempPluginBinary(t, srcDir, "roamresearch")

	installCmd := newOnePasswordCmd().Commands()[0]
	opOpts.from = srcPath
	opOpts.force = false

	buf := &strings.Builder{}
	installCmd.SetOut(buf)

	err := installCmd.RunE(installCmd, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify destination exists and is executable.
	destPath := filepath.Join(testOPPluginDir, "roamresearch")
	info, err := os.Stat(destPath)
	if err != nil {
		t.Fatalf("destination not found: %v", err)
	}
	if info.Mode()&0o100 == 0 {
		t.Error("destination binary is not executable")
	}

	expectedOutput := "Installed 1Password shell plugin: ~/.op/plugins/local/roamresearch\n" +
		"Next steps:\n" +
		"  op plugin list | grep roam-cli\n" +
		"  op plugin init roam-cli\n" +
		"  source ~/.config/op/plugins.sh\n"
	if output := buf.String(); output != expectedOutput {
		t.Errorf("unexpected output:\n%s", output)
	}
}

func TestOnePasswordInstall_SetsLocalPluginDirectoryPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	opDir := filepath.Join(tmpDir, ".op")
	pluginsDir := filepath.Join(opDir, "plugins")
	pluginDir := filepath.Join(pluginsDir, "local")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{opDir, pluginsDir, pluginDir} {
		if err := os.Chmod(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	old := testOPPluginDir
	testOPPluginDir = pluginDir
	defer func() { testOPPluginDir = old }()
	defer mockOpLookup()()

	srcDir := t.TempDir()
	srcPath := writeTempPluginBinary(t, srcDir, "roamresearch")

	installCmd := newOnePasswordCmd().Commands()[0]
	opOpts.from = srcPath
	opOpts.force = false

	if err := installCmd.RunE(installCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, dir := range []string{opDir, pluginsDir, pluginDir} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatal(err)
		}
		if got := info.Mode().Perm(); got != 0700 {
			t.Errorf("%s mode = %o, want 700", dir, got)
		}
	}
}

func TestOnePasswordInstall_WithForceOverwrites(t *testing.T) {
	_, cleanup := setupTestPluginDir(t)
	defer cleanup()
	defer mockOpLookup()()

	// Create existing plugin.
	destPath := filepath.Join(testOPPluginDir, "roamresearch")
	if err := os.WriteFile(destPath, []byte("old-contents"), 0644); err != nil {
		t.Fatal(err)
	}

	srcDir := t.TempDir()
	srcPath := writeTempPluginBinary(t, srcDir, "roamresearch")

	installCmd := newOnePasswordCmd().Commands()[0]
	opOpts.from = srcPath
	opOpts.force = true

	err := installCmd.RunE(installCmd, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(destPath)
	if err != nil {
		t.Fatalf("destination not found after overwrite: %v", err)
	}
	if info.Mode()&0o100 == 0 {
		t.Error("destination binary is not executable")
	}
}

func TestOnePasswordInstall_WithForceRejectsSymlinkDestination(t *testing.T) {
	_, cleanup := setupTestPluginDir(t)
	defer cleanup()
	defer mockOpLookup()()

	targetPath := filepath.Join(t.TempDir(), "target")
	if err := os.WriteFile(targetPath, []byte("target"), 0644); err != nil {
		t.Fatal(err)
	}
	destPath := filepath.Join(testOPPluginDir, "roamresearch")
	if err := os.Symlink(targetPath, destPath); err != nil {
		t.Fatal(err)
	}

	srcDir := t.TempDir()
	srcPath := writeTempPluginBinary(t, srcDir, "roamresearch")

	installCmd := newOnePasswordCmd().Commands()[0]
	opOpts.from = srcPath
	opOpts.force = true

	err := installCmd.RunE(installCmd, nil)
	if err == nil {
		t.Fatal("expected error for symlink destination")
	}
	if !strings.Contains(err.Error(), "symlink") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolvePluginSource_WithFromFlag(t *testing.T) {
	dir := t.TempDir()
	path := writeTempPluginBinary(t, dir, "roamresearch")

	result, err := resolvePluginSource(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != path {
		t.Errorf("expected %q, got %q", path, result)
	}
}

func TestResolvePluginSource_AcceptsAnyExecuteBit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "roamresearch")
	if err := os.WriteFile(path, []byte("#!/bin/sh\necho mock-plugin\n"), 0010); err != nil {
		t.Fatal(err)
	}

	result, err := resolvePluginSource(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != path {
		t.Errorf("expected %q, got %q", path, result)
	}
}

func TestResolvePluginSource_NonexistentFrom(t *testing.T) {
	_, err := resolvePluginSource("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for nonexistent --from path")
	}
}

func TestOnePasswordHelp(t *testing.T) {
	cmd := newOnePasswordCmd()
	buf := &strings.Builder{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	if !strings.Contains(output, "install") {
		t.Errorf("help output missing install subcommand: %s", output)
	}
}

func TestFindOpBinary_MissingInPATH(t *testing.T) {
	orig := execLookPath
	execLookPath = func(name string) (string, error) {
		return "", os.ErrNotExist
	}
	defer func() { execLookPath = orig }()

	_, err := findOpBinary()
	if err == nil {
		t.Fatal("expected error when op is not in PATH")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("unexpected error: %v", err)
	}
}
