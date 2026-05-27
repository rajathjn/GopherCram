package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_VerboseFlag(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "a.go"), []byte("package a\n"), 0o644)
	var out, errb bytes.Buffer
	code := Run(RunOptions{
		Argv:   []string{"--verbose", "--stdout", dir},
		Cwd:    dir,
		Stdout: &out, Stderr: &errb,
	})
	if code != ExitOK {
		t.Errorf("code=%v err=%s", code, errb.String())
	}
}

func TestRun_QuietSuppressesReport(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "a.go"), []byte("package a\n"), 0o644)
	out := filepath.Join(dir, "out.xml")
	var stdout, stderr bytes.Buffer
	code := Run(RunOptions{
		Argv:   []string{"--quiet", "--output", out, dir},
		Cwd:    dir,
		Stdout: &stdout, Stderr: &stderr,
	})
	if code != ExitOK {
		t.Errorf("code=%v err=%s", code, stderr.String())
	}
	if strings.Contains(stderr.String(), "Summary:") {
		t.Errorf("--quiet should suppress summary, got: %s", stderr.String())
	}
}

func TestRun_NonexistentRoot(t *testing.T) {
	var out, errb bytes.Buffer
	code := Run(RunOptions{
		Argv:   []string{"--stdout", "/nope/does/not/exist"},
		Cwd:    t.TempDir(),
		Stdout: &out, Stderr: &errb,
	})
	if code != ExitError {
		t.Errorf("expected error for missing dir, got %v", code)
	}
}

func TestRun_CopyFlag(t *testing.T) {
	// We don't assert success of the clipboard write — it depends on the
	// host — but we make sure the path doesn't crash the CLI.
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "a.go"), []byte("package a\n"), 0o644)
	var out, errb bytes.Buffer
	code := Run(RunOptions{
		Argv:   []string{"--copy", "--stdout", dir},
		Cwd:    dir,
		Stdout: &out, Stderr: &errb,
	})
	if code != ExitOK {
		t.Errorf("code=%v err=%s", code, errb.String())
	}
}
