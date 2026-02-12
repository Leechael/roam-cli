//go:build bdd

package steps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"
)

type cliState struct {
	server    *httptest.Server
	stdout    string
	stderr    string
	exitErr   error
	env       []string
	baseURL   string
	moduleDir string
}

func (s *cliState) aMockRoamAPIServer() error {
	handler := http.NewServeMux()
	handler.HandleFunc("/api/graph/testgraph/q", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&map[string]any{})
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":[["Hello BDD"]]}`))
	})
	handler.HandleFunc("/api/graph/testgraph/write", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	s.server = httptest.NewServer(handler)
	s.baseURL = s.server.URL + "/api/graph"
	return nil
}

func (s *cliState) roamCLIEnvIsConfiguredForTheMockServer() error {
	s.env = append(os.Environ(),
		"ROAM_API_TOKEN=test-token",
		"ROAM_API_GRAPH=testgraph",
		"ROAM_API_BASE_URL="+s.baseURL,
	)
	root, err := findModuleRoot()
	if err != nil {
		return err
	}
	s.moduleDir = root
	return nil
}

func (s *cliState) iRunCommand(command string) error {
	args := parseCommand(command)
	cmdArgs := append([]string{"run", "./cmd/roam-cli"}, args...)
	cmd := exec.Command("go", cmdArgs...)
	cmd.Dir = s.moduleDir
	cmd.Env = s.env
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	s.exitErr = cmd.Run()
	s.stdout = outBuf.String()
	s.stderr = errBuf.String()
	return nil
}

func (s *cliState) theCommandShouldSucceed() error {
	if s.exitErr != nil {
		return fmt.Errorf("command failed: %v, stderr: %s", s.exitErr, s.stderr)
	}
	return nil
}

func (s *cliState) stdoutShouldContain(text string) error {
	if !strings.Contains(s.stdout, text) {
		return fmt.Errorf("stdout does not contain %q, got: %s", text, s.stdout)
	}
	return nil
}

func (s *cliState) reset(*godog.Scenario) {
	s.stdout = ""
	s.stderr = ""
	s.exitErr = nil
	if s.server != nil {
		s.server.Close()
		s.server = nil
	}
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	state := &cliState{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		state.reset(sc)
		return ctx, nil
	})
	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		if state.server != nil {
			state.server.Close()
			state.server = nil
		}
		return ctx, nil
	})

	ctx.Step(`^a mock Roam API server$`, state.aMockRoamAPIServer)
	ctx.Step(`^roam-cli env is configured for the mock server$`, state.roamCLIEnvIsConfiguredForTheMockServer)
	ctx.Step(`^I run command "([^"]*)"$`, state.iRunCommand)
	ctx.Step(`^the command should succeed$`, state.theCommandShouldSucceed)
	ctx.Step(`^stdout should contain "([^"]*)"$`, state.stdoutShouldContain)
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../features"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatalf("bdd suite failed")
	}
}

func parseCommand(command string) []string {
	command = strings.TrimSpace(command)
	if strings.HasPrefix(command, "q ") {
		return []string{"q", strings.TrimSpace(strings.TrimPrefix(command, "q "))}
	}
	return strings.Fields(command)
}

func findModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}
