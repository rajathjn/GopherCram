package fileselect

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/rajathjn/GopherCram/internal/config"
)

func mkTree(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for path, content := range files {
		full := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestWalk_Basic(t *testing.T) {
	root := t.TempDir()
	mkTree(t, root, map[string]string{
		"src/main.go":    "package main",
		"src/util.go":    "package main",
		"docs/readme.md": "# hi",
		".git/HEAD":      "ignored",
		"node_modules/x": "ignored",
	})
	opts := Options{
		Roots:              []string{root},
		UseDefaultPatterns: true,
		UseGitignore:       true,
	}
	got, err := Walk(opts)
	if err != nil {
		t.Fatal(err)
	}
	gotPaths := pathsFromResults(got)
	sort.Strings(gotPaths)
	want := []string{"docs/readme.md", "src/main.go", "src/util.go"}
	if strings.Join(gotPaths, ",") != strings.Join(want, ",") {
		t.Errorf("got %v, want %v", gotPaths, want)
	}
}

func TestWalk_Include(t *testing.T) {
	root := t.TempDir()
	mkTree(t, root, map[string]string{
		"a.go":  "x",
		"a.ts":  "x",
		"b.go":  "x",
		"b.txt": "x",
	})
	got, err := Walk(Options{Roots: []string{root}, Include: []string{"*.go"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 .go files, got %d (%v)", len(got), pathsFromResults(got))
	}
}

func TestWalk_GitIgnore(t *testing.T) {
	root := t.TempDir()
	mkTree(t, root, map[string]string{
		".gitignore": "*.log\n",
		"a.go":       "x",
		"b.log":      "x",
		"c.txt":      "x",
	})
	got, err := Walk(Options{Roots: []string{root}, UseGitignore: true, UseDefaultPatterns: true})
	if err != nil {
		t.Fatal(err)
	}
	paths := pathsFromResults(got)
	for _, p := range paths {
		if strings.HasSuffix(p, ".log") {
			t.Errorf(".log file %q should be ignored", p)
		}
	}
}

func TestWalk_DotIgnore(t *testing.T) {
	root := t.TempDir()
	mkTree(t, root, map[string]string{
		".gophercramignore": "*.tmp\n",
		"a.go":              "x",
		"b.tmp":             "x",
	})
	got, err := Walk(Options{Roots: []string{root}, UseDotIgnore: true, UseDefaultPatterns: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range got {
		if strings.HasSuffix(r.RelPath, ".tmp") {
			t.Errorf("%s should be excluded by .gophercramignore", r.RelPath)
		}
	}
}

func TestWalk_CustomIgnore(t *testing.T) {
	root := t.TempDir()
	mkTree(t, root, map[string]string{
		"keep.go": "x",
		"drop.go": "x",
	})
	got, err := Walk(Options{Roots: []string{root}, CustomIgnores: []string{"drop.go"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range got {
		if r.RelPath == "drop.go" {
			t.Error("custom ignore did not apply")
		}
	}
}

func TestWalk_MaxFileSize(t *testing.T) {
	root := t.TempDir()
	big := strings.Repeat("x", 4096)
	mkTree(t, root, map[string]string{
		"big.go": big,
		"sm.go":  "x",
	})
	got, err := Walk(Options{Roots: []string{root}, MaxFileSize: 100})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range got {
		if r.RelPath == "big.go" {
			t.Error("big.go should be skipped")
		}
	}
}

func TestWalk_SingleFile(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "only.go")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Walk(Options{Roots: []string{file}})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].RelPath != "only.go" {
		t.Errorf("unexpected: %v", got)
	}
}

func TestWalk_BinaryDetectedByExt(t *testing.T) {
	root := t.TempDir()
	mkTree(t, root, map[string]string{"img.png": "fake-png"})
	got, err := Walk(Options{Roots: []string{root}})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || !got[0].IsBinary {
		t.Errorf("expected binary detection, got %+v", got)
	}
}

func TestFromConfig(t *testing.T) {
	cfg := config.Defaults()
	cfg.Include = []string{"*.go"}
	cfg.Ignore.CustomPatterns = []string{"vendor/**"}
	cfg.Input.MaxFileSize = 1000
	opts := FromConfig(&cfg, []string{"."})
	if len(opts.Include) != 1 || len(opts.CustomIgnores) != 1 || opts.MaxFileSize != 1000 {
		t.Errorf("unexpected opts: %+v", opts)
	}
}

func pathsFromResults(rs []Result) []string {
	out := make([]string, 0, len(rs))
	for _, r := range rs {
		out = append(out, r.RelPath)
	}
	return out
}
