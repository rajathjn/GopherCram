package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initLocalRepo bootstraps a fresh git repo at dir with `seed` files committed.
func initLocalRepo(t *testing.T, dir string, seed map[string]string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	for _, args := range [][]string{
		{"git", "init", "-q"},
		{"git", "config", "user.email", "t@example.test"},
		{"git", "config", "user.name", "Test"},
		{"git", "config", "commit.gpgsign", "false"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", args, err, out)
		}
	}
	for rel, content := range seed {
		full := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{
		{"git", "add", "-A"},
		{"git", "commit", "-q", "-m", "init"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", args, err, out)
		}
	}
}

func TestRunRemote_TrustConfig(t *testing.T) {
	src := t.TempDir()
	initLocalRepo(t, src, map[string]string{
		"main.go":                  "package main\n",
		"gophercram.config.json":   `{"output":{"style":"markdown"}}`,
	})
	dest := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := Run(RunOptions{
		Argv:   []string{"--remote", src, "--remote-trust-config", "--stdout"},
		Cwd:    dest,
		Stdout: &stdout, Stderr: &stderr,
	})
	if code != ExitOK {
		t.Fatalf("code=%v stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "## Files") {
		t.Errorf("expected markdown output from trusted config: %s", trunc(stdout.String(), 200))
	}
}

func TestRunRemote_CloneFailure(t *testing.T) {
	var out, errb bytes.Buffer
	code := Run(RunOptions{
		Argv:   []string{"--remote", "/this/path/does/not/exist", "--quiet"},
		Cwd:    t.TempDir(),
		Stdout: &out, Stderr: &errb,
	})
	if code != ExitError {
		t.Errorf("expected error from failing clone, got code=%v", code)
	}
}
