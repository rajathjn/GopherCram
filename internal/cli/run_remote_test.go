package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunRemote_LocalRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	src := t.TempDir()
	for _, args := range [][]string{
		{"git", "init", "-q"},
		{"git", "config", "user.email", "t@example.test"},
		{"git", "config", "user.name", "Test"},
		{"git", "config", "commit.gpgsign", "false"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = src
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", args, err, out)
		}
	}
	if err := os.WriteFile(filepath.Join(src, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"git", "add", "main.go"},
		{"git", "commit", "-q", "-m", "init"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = src
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", args, err, out)
		}
	}
	dest := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := Run(RunOptions{
		Argv:   []string{"--remote", src, "--stdout", "--style", "json"},
		Cwd:    dest,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if code != ExitOK {
		t.Fatalf("code=%v stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "main.go") {
		t.Errorf("cloned repo output missing main.go: %s", trunc(stdout.String(), 200))
	}
}

func TestRunRemote_BadURL(t *testing.T) {
	var out, errb bytes.Buffer
	code := Run(RunOptions{
		Argv:   []string{"--remote", ""},
		Stdout: &out, Stderr: &errb,
	})
	if code != ExitError {
		t.Errorf("expected error for empty remote, got code=%v", code)
	}
}

func TestRunRemoteAutoDetect(t *testing.T) {
	// `IsExplicitRemoteURL` is exercised by gitops; here we just make sure the
	// CLI doesn't accidentally treat a positional argument that *looks* like a
	// shorthand but is actually a local directory as remote when --remote
	// isn't passed.
	dir := t.TempDir()
	var out, errb bytes.Buffer
	code := Run(RunOptions{
		Argv:   []string{"--stdout", dir},
		Cwd:    dir,
		Stdout: &out, Stderr: &errb,
	})
	if code != ExitOK {
		t.Errorf("code=%v stderr=%s", code, errb.String())
	}
}
