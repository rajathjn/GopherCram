package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runWithFixture(t *testing.T, args []string) (string, string, ExitCode) {
	t.Helper()
	dir := t.TempDir()
	mk := func(rel, content string) {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mk("src/main.go", "package main\n\nfunc main() {}\n")
	mk("src/util.go", "package main\n")
	mk("README.md", "# project\n")
	mk(".gitignore", "*.log\n")

	var out, errb bytes.Buffer
	code := Run(RunOptions{
		Argv:   append(args, dir),
		Cwd:    dir,
		Stdin:  strings.NewReader(""),
		Stdout: &out,
		Stderr: &errb,
	})
	return out.String(), errb.String(), code
}

func TestRunVersion(t *testing.T) {
	var out, errb bytes.Buffer
	if code := Run(RunOptions{Argv: []string{"--version"}, Stdout: &out, Stderr: &errb}); code != ExitOK {
		t.Errorf("code=%v", code)
	}
	if !strings.Contains(out.String(), AppName) {
		t.Errorf("output=%q", out.String())
	}
}

func TestRunHelp(t *testing.T) {
	var out, errb bytes.Buffer
	if code := Run(RunOptions{Argv: []string{"--help"}, Stdout: &out, Stderr: &errb}); code != ExitOK {
		t.Errorf("code=%v", code)
	}
	if !strings.Contains(out.String(), "Usage:") {
		t.Errorf("missing usage in help")
	}
}

func TestRunUnknownFlag(t *testing.T) {
	var out, errb bytes.Buffer
	if code := Run(RunOptions{Argv: []string{"--no-such-flag"}, Stdout: &out, Stderr: &errb}); code != ExitUsage {
		t.Errorf("code=%v", code)
	}
}

func TestRunInitFlag(t *testing.T) {
	dir := t.TempDir()
	var out, errb bytes.Buffer
	if code := Run(RunOptions{Argv: []string{"--init"}, Cwd: dir, Stdout: &out, Stderr: &errb}); code != ExitOK {
		t.Errorf("code=%v err=%s", code, errb.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "gophercram.config.json")); err != nil {
		t.Errorf("config not written: %v", err)
	}
}

func TestRunPackXML(t *testing.T) {
	out, errb, code := runWithFixture(t, []string{"--stdout"})
	if code != ExitOK {
		t.Fatalf("code=%v err=%s", code, errb)
	}
	if !strings.Contains(out, "<gophercram>") || !strings.Contains(out, "src/main.go") {
		t.Errorf("unexpected XML: %s", trunc(out, 200))
	}
}

func TestRunPackMarkdown(t *testing.T) {
	out, errb, code := runWithFixture(t, []string{"--stdout", "--style", "markdown"})
	if code != ExitOK {
		t.Fatalf("code=%v err=%s", code, errb)
	}
	if !strings.Contains(out, "## Files") {
		t.Errorf("expected markdown structure: %s", trunc(out, 200))
	}
}

func TestRunPackJSON(t *testing.T) {
	out, errb, code := runWithFixture(t, []string{"--stdout", "--style", "json"})
	if code != ExitOK {
		t.Fatalf("code=%v err=%s", code, errb)
	}
	if !strings.Contains(out, `"files":`) {
		t.Errorf("expected json files key: %s", trunc(out, 200))
	}
}

func TestRunPackPlain(t *testing.T) {
	out, _, code := runWithFixture(t, []string{"--stdout", "--style", "plain"})
	if code != ExitOK {
		t.Fatalf("code=%v", code)
	}
	if !strings.Contains(out, "SUMMARY") {
		t.Errorf("expected plain output: %s", trunc(out, 200))
	}
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func TestRunStdinPaths(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.txt")
	_ = os.WriteFile(file, []byte("hello"), 0o644)
	var out, errb bytes.Buffer
	code := Run(RunOptions{
		Argv:   []string{"--stdin", "--stdout"},
		Cwd:    dir,
		Stdin:  strings.NewReader(file + "\n"),
		Stdout: &out, Stderr: &errb,
	})
	if code != ExitOK {
		t.Fatalf("code=%v err=%s", code, errb.String())
	}
	if !strings.Contains(out.String(), "a.txt") {
		t.Errorf("output=%q", out.String())
	}
}

func TestRunStdinEmpty(t *testing.T) {
	var out, errb bytes.Buffer
	code := Run(RunOptions{
		Argv: []string{"--stdin", "--stdout"}, Cwd: t.TempDir(),
		Stdin: strings.NewReader(""), Stdout: &out, Stderr: &errb,
	})
	if code != ExitError {
		t.Errorf("expected error when stdin is empty, got %v", code)
	}
}

func TestRunWriteToDisk(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "x.go"), []byte("package x\n"), 0o644)
	outFile := filepath.Join(dir, "out.xml")
	var out, errb bytes.Buffer
	code := Run(RunOptions{
		Argv: []string{"--output", outFile, "--quiet", dir},
		Cwd:  dir, Stdout: &out, Stderr: &errb,
	})
	if code != ExitOK {
		t.Fatalf("code=%v err=%s", code, errb.String())
	}
	body, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "<gophercram>") {
		t.Errorf("output not written as XML")
	}
}

func TestRunSecurityScan(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "leaks.txt"),
		[]byte("aws=AKIAIOSFODNN7EXAMPLE\n"), 0o644)
	var out, errb bytes.Buffer
	code := Run(RunOptions{
		Argv: []string{dir}, Cwd: dir,
		Stdout: &out, Stderr: &errb,
	})
	if code != ExitOK {
		t.Fatalf("code=%v err=%s", code, errb.String())
	}
	if !strings.Contains(errb.String(), "Security") {
		t.Errorf("expected security report in stderr: %q", errb.String())
	}
}

func TestRunConfigFile(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "a.go"), []byte("package a\n"), 0o644)
	cfg := `{"output": {"style": "markdown"}}`
	_ = os.WriteFile(filepath.Join(dir, "gophercram.config.json"), []byte(cfg), 0o644)
	var out, errb bytes.Buffer
	code := Run(RunOptions{
		Argv:   []string{"--stdout", dir},
		Cwd:    dir,
		Stdout: &out, Stderr: &errb,
	})
	if code != ExitOK {
		t.Fatalf("code=%v err=%s", code, errb.String())
	}
	if !strings.Contains(out.String(), "## Files") {
		t.Errorf("config file style not applied")
	}
}

func TestRunBadFlag(t *testing.T) {
	var out, errb bytes.Buffer
	code := Run(RunOptions{Argv: []string{"--split-output", "garbage"}, Stdout: &out, Stderr: &errb})
	if code != ExitError {
		t.Errorf("expected error, got %v", code)
	}
}

func TestReadStdinPaths(t *testing.T) {
	paths, err := readStdinPaths(strings.NewReader("file1\n# comment\n\nfile2\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 2 || paths[0] != "file1" || paths[1] != "file2" {
		t.Errorf("paths=%v", paths)
	}
}
