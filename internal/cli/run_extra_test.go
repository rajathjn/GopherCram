package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunTokenCountTree_Integration(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "x.go"), []byte(strings.Repeat("var x = 1\n", 200)), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "y.go"), []byte("var y = 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	code := Run(RunOptions{
		Argv:   []string{"--token-count-tree", "10", "--stdout", "--style", "json", dir},
		Cwd:    dir,
		Stdout: &out, Stderr: &errb,
	})
	if code != ExitOK {
		t.Fatalf("code=%v stderr=%s", code, errb.String())
	}
	if !strings.Contains(out.String(), "Token-count tree:") {
		t.Errorf("token tree not printed: %s", trunc(out.String(), 400))
	}
}

func TestRunTopFilesLen(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 8; i++ {
		_ = os.WriteFile(filepath.Join(dir, "f"+string(rune('a'+i))+".go"),
			[]byte("package main\n"), 0o644)
	}
	var out, errb bytes.Buffer
	code := Run(RunOptions{
		Argv:   []string{"--top-files-len", "3", dir},
		Cwd:    dir,
		Stdout: &out, Stderr: &errb,
	})
	if code != ExitOK {
		t.Fatalf("code=%v stderr=%s", code, errb.String())
	}
	if !strings.Contains(errb.String(), "Top 3 files") {
		t.Errorf("expected top-3 header: %s", errb.String())
	}
}

func TestRunConfigViaFlag(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "custom.json")
	_ = os.WriteFile(cfg, []byte(`{"output":{"style":"json"}}`), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "a.go"), []byte("package a\n"), 0o644)
	var out, errb bytes.Buffer
	code := Run(RunOptions{
		Argv:   []string{"--config", cfg, "--stdout", dir},
		Cwd:    dir,
		Stdout: &out, Stderr: &errb,
	})
	if code != ExitOK {
		t.Fatalf("code=%v stderr=%s", code, errb.String())
	}
	if !strings.Contains(out.String(), `"files":`) {
		t.Errorf("style from --config not applied: %s", trunc(out.String(), 200))
	}
}

func TestRunBadConfigFile(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "bad.json")
	_ = os.WriteFile(cfg, []byte(`{not json`), 0o644)
	var out, errb bytes.Buffer
	code := Run(RunOptions{
		Argv: []string{"--config", cfg, dir},
		Cwd:  dir, Stdout: &out, Stderr: &errb,
	})
	if code != ExitError {
		t.Errorf("expected error for bad config, got %v", code)
	}
}

func TestRunOutputDash(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "a.go"), []byte("package a\n"), 0o644)
	var out, errb bytes.Buffer
	code := Run(RunOptions{
		Argv: []string{"--output", "-", dir},
		Cwd:  dir, Stdout: &out, Stderr: &errb,
	})
	if code != ExitOK {
		t.Fatalf("code=%v stderr=%s", code, errb.String())
	}
	if !strings.Contains(out.String(), "<gophercram>") {
		t.Errorf("expected stdout output: %s", trunc(out.String(), 200))
	}
}
