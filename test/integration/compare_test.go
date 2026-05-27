// Package integration contains end-to-end tests that compare GopherCram's
// output against the upstream Node-based repomix CLI invoked via `npx`.
//
// The tests are gated on `npx` being available on PATH. Set
// GOPHERCRAM_INTEGRATION=1 to enable them; otherwise they skip.
package integration

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/rajathjn/GopherCram/internal/cli"
)

// fixturePath returns the absolute path to the bundled dummy TS repo.
func fixturePath(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not resolve test file path")
	}
	return filepath.Join(filepath.Dir(file), "..", "fixtures", "dummy-ts-repo")
}

// requireOptIn skips the test unless the integration env var is set.
func requireOptIn(t *testing.T) {
	if os.Getenv("GOPHERCRAM_INTEGRATION") != "1" {
		t.Skip("set GOPHERCRAM_INTEGRATION=1 to run integration tests")
	}
	if _, err := exec.LookPath("npx"); err != nil {
		t.Skip("npx not installed")
	}
}

func runGopherCram(t *testing.T, fixture string, extra ...string) string {
	t.Helper()
	tmp := t.TempDir()
	out := filepath.Join(tmp, "gc.json")
	argv := append([]string{
		"--style", "json",
		"--output", out,
		"--quiet",
		fixture,
	}, extra...)
	var stderr bytes.Buffer
	code := cli.Run(cli.RunOptions{
		Argv:   argv,
		Cwd:    tmp,
		Stderr: &stderr,
		Stdout: &bytes.Buffer{},
	})
	if code != cli.ExitOK {
		t.Fatalf("gophercram failed (%d): %s", code, stderr.String())
	}
	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func runRepomix(t *testing.T, fixture string, extra ...string) string {
	t.Helper()
	tmp := t.TempDir()
	out := filepath.Join(tmp, "rmx.txt")
	argv := append([]string{
		"--yes", "repomix",
		"--style", "plain",
		"--output", out,
		fixture,
	}, extra...)
	cmd := exec.Command("npx", argv...)
	cmd.Dir = tmp
	cmd.Env = append(os.Environ(), "REPOMIX_NO_TELEMETRY=1")
	var combined bytes.Buffer
	cmd.Stdout = &combined
	cmd.Stderr = &combined
	if err := cmd.Run(); err != nil {
		t.Skipf("npx repomix not runnable in this environment: %v\n%s", err, combined.String())
	}
	body, err := os.ReadFile(out)
	if err != nil {
		t.Skipf("repomix produced no output: %v", err)
	}
	// Append the captured progress report so callers can also grep token /
	// summary lines.
	return string(body) + "\n----PROGRESS----\n" + combined.String()
}

// expectedFiles is the set of paths we expect any sane pack tool to include
// when run against the fixture (with default gitignore + default patterns).
// Files outside this list are either gitignored (.env, *.log) or covered by
// the default ignore patterns (node_modules, dist).
var expectedFiles = []string{
	"README.md",
	"package.json",
	"tsconfig.json",
	".env.example",
	".gitignore",
	"src/index.ts",
	"src/api/users.ts",
	"src/api/products.ts",
	"src/utils/format.ts",
	"src/utils/validate.ts",
	"src/components/Button.tsx",
	"src/components/Modal.tsx",
	"src/config.ts",
	"tests/api.test.ts",
	"tests/utils.test.ts",
	"docs/README.md",
}

// excludedFiles must NOT appear in either tool's output.
var excludedFiles = []string{
	".env",
	"build.log",
	"node_modules/fake-dep/index.js",
}

func TestGopherCram_FilesIncluded(t *testing.T) {
	fixture := fixturePath(t)
	body := runGopherCram(t, fixture)
	for _, p := range expectedFiles {
		if !strings.Contains(body, p) {
			t.Errorf("expected %s in GopherCram output", p)
		}
	}
}

func TestGopherCram_FilesExcluded(t *testing.T) {
	fixture := fixturePath(t)
	body := runGopherCram(t, fixture)
	// `.env` is gitignored, `build.log` matches *.log gitignore, node_modules
	// is filtered by the default patterns.
	for _, p := range excludedFiles {
		if strings.Contains(body, "\""+p+"\"") {
			t.Errorf("did not expect %s in GopherCram output", p)
		}
	}
}

func TestGopherCram_SecurityScanner(t *testing.T) {
	fixture := fixturePath(t)
	body := runGopherCram(t, fixture)
	var doc map[string]any
	if err := json.Unmarshal([]byte(body), &doc); err != nil {
		t.Fatal(err)
	}
	// src/config.ts contains the AWS access key and GitHub token. The packager
	// removes the content but keeps the path with a `skipped` reason.
	files, _ := doc["files"].([]any)
	var found bool
	for _, f := range files {
		m, _ := f.(map[string]any)
		if m["path"] == "src/config.ts" {
			if reason, ok := m["skipped"].(string); ok && strings.Contains(reason, "security") {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected src/config.ts to be flagged + skipped by the security scanner")
	}
}

func TestCompare_FileCoverage(t *testing.T) {
	requireOptIn(t)
	fixture := fixturePath(t)

	start := time.Now()
	gc := runGopherCram(t, fixture)
	t.Logf("gophercram took %s", time.Since(start))

	start = time.Now()
	rm := runRepomix(t, fixture)
	t.Logf("npx repomix took %s", time.Since(start))

	// Both outputs must mention the same set of source files (modulo the
	// content-removed ones — repomix and GopherCram disagree slightly on
	// which secret files to omit, but both should report .env was ignored).
	for _, p := range []string{
		"src/index.ts",
		"src/api/users.ts",
		"src/utils/format.ts",
		"tests/api.test.ts",
	} {
		if !strings.Contains(gc, p) {
			t.Errorf("GopherCram missing %s", p)
		}
		if !strings.Contains(rm, p) {
			t.Errorf("repomix missing %s", p)
		}
	}
	// gitignore rules respected by both tools — check by looking for the
	// secret content of `.env`, not the literal substring `.env` (which
	// matches `.env.example`).
	for _, leak := range []string{
		"SESSION_SECRET=this-secret-stays-out",
		"wJalrXUtnFEMIK7MDENGbPxRfiCYEXAMPLEKEY",
	} {
		if strings.Contains(rm, leak) {
			t.Errorf("repomix leaked .env content: %q", leak)
		}
		if strings.Contains(gc, leak) {
			t.Errorf("GopherCram leaked .env content: %q", leak)
		}
	}
	if strings.Contains(rm, "build.log\n[build]") {
		t.Error("repomix did not ignore build.log")
	}
}

// TestCompare_TokenCountSimilar checks that GopherCram's approximate token
// count is within ±25% of repomix's tiktoken-based count over the fixture.
// We can't expect bit-exact agreement (different algorithms), but staying
// within a reasonable band is a good guard against drift.
func TestCompare_TokenCountSimilar(t *testing.T) {
	requireOptIn(t)
	fixture := fixturePath(t)

	gc := runGopherCram(t, fixture)
	var doc map[string]any
	if err := json.Unmarshal([]byte(gc), &doc); err != nil {
		t.Fatal(err)
	}
	stats, _ := doc["stats"].(map[string]any)
	gcTokens, _ := stats["tokens"].(float64)
	if gcTokens <= 0 {
		t.Fatalf("gophercram reported %v tokens", gcTokens)
	}

	rmx := runRepomix(t, fixture)
	// repomix's progress report contains a line like
	// "Total Tokens: 2,737 tokens". We skip past the colon, then collect
	// digits (ignoring commas) until we hit whitespace.
	idx := strings.Index(rmx, "Total Tokens")
	if idx < 0 {
		t.Skip("could not locate repomix token report")
	}
	line := rmx[idx:]
	if cut := strings.IndexByte(line, '\n'); cut >= 0 {
		line = line[:cut]
	}
	if colon := strings.IndexByte(line, ':'); colon >= 0 {
		line = line[colon+1:]
	}
	var rmTokens int
	started := false
	for _, r := range line {
		if r >= '0' && r <= '9' {
			rmTokens = rmTokens*10 + int(r-'0')
			started = true
		} else if r == ',' {
			continue
		} else if started {
			break
		}
	}
	if rmTokens <= 0 {
		t.Skipf("could not parse repomix tokens from %q", line)
	}
	ratio := float64(gcTokens) / float64(rmTokens)
	if ratio < 0.75 || ratio > 1.25 {
		t.Errorf("token estimates diverge too much: gc=%v rm=%d (ratio %.2f)", gcTokens, rmTokens, ratio)
	} else {
		t.Logf("token estimates agree: gc=%v rm=%d (ratio %.2f)", gcTokens, rmTokens, ratio)
	}
}
