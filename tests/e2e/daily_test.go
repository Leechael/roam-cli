//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

var (
	keepPages      = flag.Bool("keep-pages", envBool("DAILY_USE_SMOKE_KEEP_PAGES"), "keep scratch pages after e2e test failures")
	requestTimeout = flag.Int("roam-timeout", envInt("DAILY_USE_SMOKE_TIMEOUT_SECONDS", 60), "roam-cli request timeout in seconds")

	repoRoot string
	tmpDir   string
	cliPath  string
	runID    string
)

func TestMain(m *testing.M) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Fprintln(os.Stderr, "failed to resolve test file path")
		os.Exit(1)
	}
	repoRoot = filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))

	loadEnvFiles()

	var err error
	tmpDir, err = os.MkdirTemp("", "roam-cli-e2e-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	if override := strings.TrimSpace(os.Getenv("ROAM_CLI")); override != "" {
		cliPath = override
	} else {
		cliPath = filepath.Join(tmpDir, "roam-cli")
		cmd := exec.Command("go", "build", "-o", cliPath, "./cmd/roam-cli")
		cmd.Dir = repoRoot
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to build roam-cli: %v\n%s\n", err, out)
			os.Exit(1)
		}
	}

	runID = fmt.Sprintf("%d-%d", time.Now().Unix(), os.Getpid())
	os.Exit(m.Run())
}

func TestDailyUse(t *testing.T) {
	e := newE2E(t)

	t.Run("Status", func(t *testing.T) {
		c := e.newCase(t)
		c.mustRun("status")
	})

	t.Run("PageDelete", func(t *testing.T) {
		c := e.newCase(t)
		page := c.newPage("delete")

		c.mustPipe("- temporary content", "save", "--title", page)
		out := c.mustRun("get", page)
		c.assertContains(out, "temporary content", "created page content was not readable before delete")
		c.mustRun("page", "delete", page)
		c.assertPageDeleted(page)
	})

	t.Run("MoveUnderInvalidUIDDoesNotPollutePage", func(t *testing.T) {
		c := e.newCase(t)
		page := c.newPage("move-invalid")
		missingSection := "[[Daily Use Smoke Should Not Exist " + runID + "]]"
		badUID := "bad-daily-use-" + runID

		c.logf("bad source uid: %s", badUID)
		c.logf("section that must not appear: %s", missingSection)
		c.mustPipe("- existing content", "save", "--title", page)
		if out, err := c.run("", "move", "--uid", badUID, "--title", page, "--under", missingSection); err == nil {
			c.failf("move --under unexpectedly succeeded with invalid source UID\n%s", out)
		}
		out := c.mustRun("get", page)
		c.assertNotContains(out, missingSection, "move --under created the section before validating the source UID")
	})

	t.Run("SavePlainReturnsUIDForNestedFollowUp", func(t *testing.T) {
		c := e.newCase(t)
		page := c.newPage("plain")
		inbox := "[[Daily Use Smoke Inbox " + runID + "]]"
		parentText := "Daily Use smoke parent " + runID
		childText := "Daily Use smoke child " + runID

		c.logf("inbox section: %s", inbox)
		parentUID := c.mustPipe("- "+parentText, "save", "--title", page, "--under", inbox, "--plain")
		parentUID = strings.TrimSpace(parentUID)
		c.logf("save --plain returned uid: %s", parentUID)
		c.mustPipe("- "+childText, "save", "--parent", parentUID)
		out := c.mustRun("get", page)
		if !hasExactLine(out, "    "+childText) {
			c.failf("save --plain did not return the new parent block UID for nested follow-up content")
		}
		out = c.mustRun("get", parentUID)
		c.assertContains(out, parentText, "get block UID did not include saved parent text")
		c.assertContains(out, childText, "get block UID did not include nested child text")
	})

	t.Run("SaveReplaceReplacesNamedPageContent", func(t *testing.T) {
		c := e.newCase(t)
		page := c.newPage("replace")
		oldText := "Daily Use old replace content " + runID
		newText := "Daily Use new replace content " + runID

		c.mustPipe("- "+oldText, "save", "--title", page)
		c.mustPipe("- "+newText, "save", "--title", page, "--replace")
		out := c.mustRun("get", page)
		c.assertContains(out, newText, "replace page is missing new content")
		c.assertNotContains(out, oldText, "replace page still contains old content")
	})

	t.Run("SearchPageAndBlockModesFindScratchContent", func(t *testing.T) {
		c := e.newCase(t)
		page := c.newPage("search")
		searchToken := "DailyUseSearchToken" + runID

		c.mustPipe("- "+searchToken+" page result", "save", "--title", page)
		out := c.mustRun("search", searchToken, "--type", "page", "--limit", "1")
		c.assertContains(out, page, "search --type page did not include scratch page")
		out = c.mustRun("search", searchToken, "--type", "block", "--page", page, "--limit", "1")
		c.assertContains(out, searchToken, "search --type block did not include scratch block text")
	})

	t.Run("MoveExistingBlockToNamedPageSection", func(t *testing.T) {
		c := e.newCase(t)
		sourcePage := c.newPage("move-source")
		targetPage := c.newPage("move-target")
		moveText := "Daily Use moved block " + runID
		moveSection := "[[Daily Use Move Target " + runID + "]]"

		moveUID := strings.TrimSpace(c.mustPipe("- "+moveText, "save", "--title", sourcePage, "--plain"))
		c.logf("move source uid: %s", moveUID)
		c.mustPipe("- target seed", "save", "--title", targetPage)
		c.mustRun("move", "--uid", moveUID, "--title", targetPage, "--under", moveSection)
		out := c.mustRun("get", targetPage)
		c.assertContains(out, moveSection, "move target page is missing destination section")
		c.assertContains(out, moveText, "move target page is missing moved block")
	})

	t.Run("PageClearRemovesContentAndKeepsPage", func(t *testing.T) {
		c := e.newCase(t)
		page := c.newPage("clear")
		clearText := "Daily Use clear content " + runID

		c.mustPipe("- "+clearText, "save", "--title", page)
		c.mustRun("page", "clear", page)
		out := c.mustRun("get", page)
		c.assertNotContains(out, clearText, "page clear left old content behind")
	})

	t.Run("SaveUnderExistingDailySectionKeepsSectionDepth", func(t *testing.T) {
		c := e.newCase(t)
		dailyDay := (os.Getpid() % 28) + 1
		dailyDate := fmt.Sprintf("2099-12-%02d", dailyDay)
		dailyPage := dailyTitle(2099, time.December, dailyDay)
		c.addPage(dailyPage)

		underSection := "[[Daily Use Repeated Under " + runID + "]]"
		preludeText := "Daily Use daily prelude " + runID
		preludeChildText := "Daily Use daily prelude child " + runID
		preludeLeafText := "Daily Use daily prelude leaf " + runID
		firstUnderText := "Daily Use first under append " + runID
		firstChildText := "Daily Use first under child " + runID
		tailText := "Daily Use later top-level tail " + runID
		tailChildText := "Daily Use later tail child " + runID
		tailLeafText := "Daily Use later tail leaf " + runID
		secondUnderText := "Daily Use second under append " + runID
		secondChildText := "Daily Use second under child " + runID

		c.logf("daily page date: %s", dailyDate)
		c.logf("daily page title: %s", dailyPage)
		c.logf("under section: %s", underSection)

		c.createPageOnly(dailyPage)
		preludeUID := strings.TrimSpace(c.mustPipe("- "+preludeText, "save", "--to-daily-page", dailyDate, "--plain"))
		preludeChildUID := strings.TrimSpace(c.mustPipe("- "+preludeChildText, "save", "--parent", preludeUID, "--plain"))
		c.mustPipe("- "+preludeLeafText, "save", "--parent", preludeChildUID)

		firstInput := fmt.Sprintf("- %s\n  - %s", firstUnderText, firstChildText)
		firstUnderUID := strings.TrimSpace(c.mustPipe(firstInput, "save", "--to-daily-page", dailyDate, "--under", underSection, "--plain"))
		c.logf("first --under uid: %s", firstUnderUID)

		tailInput := fmt.Sprintf("- %s\n  - %s\n    - %s", tailText, tailChildText, tailLeafText)
		tailUID := strings.TrimSpace(c.mustPipe(tailInput, "save", "--title", dailyPage, "--plain"))
		c.logf("later top-level tail uid: %s", tailUID)

		secondInput := fmt.Sprintf("- %s\n  - %s", secondUnderText, secondChildText)
		secondUnderUID := strings.TrimSpace(c.mustPipe(secondInput, "save", "--to-daily-page", dailyDate, "--under", underSection, "--plain"))
		c.logf("second --under uid: %s", secondUnderUID)

		out := c.mustRun("get", dailyPage)
		c.assertContains(out, underSection, "daily page is missing repeated --under section")
		if !hasExactLine(out, "  "+firstUnderText) {
			c.failf("first --under append is not a direct child of the section")
		}
		if !hasExactLine(out, "  "+secondUnderText) {
			c.failf("second --under append is not a direct child of the existing section")
		}
		if hasLineWithMinIndent(out, secondUnderText, 4) {
			c.failf("second --under append was written into a deeper child")
		}
		if !hasExactLine(out, tailText) {
			c.failf("later tail block is not a top-level daily page block")
		}

		out = c.mustRun("get", firstUnderUID)
		c.assertNotContains(out, secondUnderText, "second --under append was nested under the first --under append")
		out = c.mustRun("get", tailUID)
		c.assertNotContains(out, secondUnderText, "second --under append was nested under the later top-level tail")
		out = c.mustRun("get", secondUnderUID)
		c.assertContains(out, secondChildText, "second --under append is missing its own child")
	})
}

type e2eEnv struct {
	cli     string
	timeout int
}

type testCase struct {
	t     *testing.T
	env   *e2eEnv
	log   bytes.Buffer
	pages []string
}

func newE2E(t *testing.T) *e2eEnv {
	t.Helper()
	missing := missingEnv("ROAM_API_TOKEN", "ROAM_API_GRAPH")
	if len(missing) > 0 {
		t.Skipf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	return &e2eEnv{cli: cliPath, timeout: *requestTimeout}
}

func (e *e2eEnv) newCase(t *testing.T) *testCase {
	t.Helper()
	c := &testCase{t: t, env: e}
	t.Cleanup(func() {
		if t.Failed() {
			t.Logf("e2e command log:\n%s", strings.TrimRight(c.log.String(), "\n"))
		}
	})
	t.Cleanup(func() {
		c.cleanupPages()
	})
	return c
}

func (c *testCase) newPage(suffix string) string {
	c.t.Helper()
	page := fmt.Sprintf("roam-cli-daily-use-smoke-%s-%s", runID, suffix)
	c.addPage(page)
	c.logf("scratch page[%s]: %s", suffix, page)
	return page
}

func (c *testCase) addPage(page string) {
	c.t.Helper()
	c.pages = append(c.pages, page)
}

func (c *testCase) createPageOnly(title string) {
	c.t.Helper()
	actions := []map[string]any{
		{
			"action": "create-page",
			"page": map[string]any{
				"title":              title,
				"uid":                newUID(),
				"children-view-type": "bullet",
			},
		},
	}
	raw, err := json.Marshal(actions)
	if err != nil {
		c.failf("marshal create-page action: %v", err)
	}
	c.mustPipe(string(raw), "batch", "run")
}

func (c *testCase) assertPageDeleted(page string) {
	c.t.Helper()
	out, err := c.run("", "get", page)
	if err == nil {
		c.failf("page still exists after page delete: %s\n%s", page, out)
	}
	if !strings.Contains(out, "not found") {
		c.failf("expected get to report not found after page delete: %s\n%s", page, out)
	}
}

func (c *testCase) mustRun(args ...string) string {
	c.t.Helper()
	out, err := c.run("", args...)
	if err != nil {
		c.failf("command failed: roam-cli %s\n%s", strings.Join(args, " "), out)
	}
	return out
}

func (c *testCase) mustPipe(input string, args ...string) string {
	c.t.Helper()
	out, err := c.run(input, args...)
	if err != nil {
		c.failf("command failed: printf input | roam-cli %s\n%s", strings.Join(args, " "), out)
	}
	return out
}

func (c *testCase) run(input string, args ...string) (string, error) {
	c.t.Helper()
	cmdArgs := append([]string{"--timeout", strconv.Itoa(c.env.timeout)}, args...)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.env.timeout+30)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.env.cli, cmdArgs...)
	cmd.Dir = repoRoot
	if input != "" {
		cmd.Stdin = strings.NewReader(input + "\n")
	}
	c.logf("+ roam-cli %s", strings.Join(cmdArgs, " "))
	if input != "" {
		c.logf("stdin:\n%s", input)
	}
	out, err := cmd.CombinedOutput()
	output := string(out)
	if ctx.Err() == context.DeadlineExceeded {
		err = ctx.Err()
	}
	if err != nil {
		c.logf("exit=error: %v", err)
	} else {
		c.logf("exit=0")
	}
	if output == "" {
		c.logf("output: <empty>")
	} else {
		c.logf("output:\n%s", strings.TrimRight(output, "\n"))
	}
	return strings.TrimRight(output, "\n"), err
}

func (c *testCase) cleanupPages() {
	if *keepPages || len(c.pages) == 0 {
		if *keepPages && len(c.pages) > 0 {
			c.logf("keep-pages enabled; leaving scratch pages for inspection: %s", strings.Join(c.pages, " "))
		}
		return
	}
	for _, page := range c.pages {
		c.logf("cleanup: page clear %s", page)
		_, _ = c.runCleanup("page", "clear", page)
		c.logf("cleanup: page delete %s", page)
		_, _ = c.runCleanup("page", "delete", page)
	}
}

func (c *testCase) runCleanup(args ...string) (string, error) {
	cmdArgs := append([]string{"--timeout", strconv.Itoa(c.env.timeout)}, args...)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.env.timeout+30)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, c.env.cli, cmdArgs...)
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		err = ctx.Err()
	}
	return string(out), err
}

func (c *testCase) assertContains(haystack, needle, message string) {
	c.t.Helper()
	if !strings.Contains(haystack, needle) {
		c.failf("%s", message)
	}
}

func (c *testCase) assertNotContains(haystack, needle, message string) {
	c.t.Helper()
	if strings.Contains(haystack, needle) {
		c.failf("%s", message)
	}
}

func (c *testCase) failf(format string, args ...any) {
	c.t.Helper()
	c.t.Fatalf(format, args...)
}

func (c *testCase) logf(format string, args ...any) {
	fmt.Fprintf(&c.log, "[%s] ", time.Now().Format("15:04:05"))
	fmt.Fprintf(&c.log, format, args...)
	c.log.WriteByte('\n')
}

func loadEnvFiles() {
	cwd, err := os.Getwd()
	if err == nil {
		_ = loadEnvFile(filepath.Join(cwd, ".env"))
	}
	rootEnv := filepath.Join(repoRoot, ".env")
	if err != nil || filepath.Clean(filepath.Join(cwd, ".env")) != rootEnv {
		_ = loadEnvFile(rootEnv)
	}
}

func loadEnvFile(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		value = strings.TrimSpace(value)
		if unquoted, err := strconv.Unquote(value); err == nil {
			value = unquoted
		}
		_ = os.Setenv(key, value)
	}
	return nil
}

func missingEnv(keys ...string) []string {
	missing := []string{}
	for _, key := range keys {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			missing = append(missing, key)
		}
	}
	return missing
}

func envBool(key string) bool {
	return os.Getenv(key) == "1" || strings.EqualFold(os.Getenv(key), "true")
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func newUID() string {
	return uuid.NewString()
}

func dailyTitle(year int, month time.Month, day int) string {
	return fmt.Sprintf("%s %d%s, %d", month.String(), day, ordinalSuffix(day), year)
}

func ordinalSuffix(day int) string {
	if day >= 11 && day <= 13 {
		return "th"
	}
	switch day % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	default:
		return "th"
	}
}

func hasExactLine(text, want string) bool {
	for _, line := range strings.Split(text, "\n") {
		if line == want {
			return true
		}
	}
	return false
}

func hasLineWithMinIndent(text, suffix string, minIndent int) bool {
	prefix := strings.Repeat(" ", minIndent)
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, prefix) && strings.TrimLeft(line, " ") == suffix {
			return true
		}
	}
	return false
}
