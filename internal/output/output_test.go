package output

import (
	"strings"
	"testing"
	"time"

	"github.com/rajathjn/GopherCram/internal/config"
	"github.com/rajathjn/GopherCram/internal/fileselect"
	"github.com/rajathjn/GopherCram/internal/metrics"
)

func sampleDoc(style config.OutputStyle) *Document {
	cfg := config.Defaults()
	cfg.Output.Style = style
	return &Document{
		GeneratedAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Config:        &cfg,
		HeaderText:    "hello header",
		Instruction:   "do the thing",
		DirectoryTree: "src/\n  main.go\n",
		Files: []fileselect.File{
			{Result: fileselect.Result{RelPath: "src/main.go", Size: 12}, Content: "package main\n"},
			{Result: fileselect.Result{RelPath: "img.png", IsBinary: true}, Skipped: true, SkipReason: "binary"},
		},
		Aggregate: metrics.Aggregate{
			TotalFiles: 1, TotalChars: 12, TotalTokens: 4, TotalBytes: 12,
		},
		GitDiffWorkTree: "diff --git a/x b/x",
		GitDiffStaged:   "",
		GitLog: []GitCommit{{
			Hash: "abc123", Author: "Tester", Date: "2026-01-01",
			Message: "init", Files: []string{"src/main.go"},
		}},
		AppName: "GopherCram", AppVersion: "test",
	}
}

func TestXMLRender(t *testing.T) {
	doc := sampleDoc(config.StyleXML)
	r := For(config.StyleXML)
	out, err := r.Render(doc)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`<?xml version="1.0"`,
		"<gophercram>",
		"<summary>",
		"<user_header>",
		"<directory_structure>",
		"src/main.go",
		"<git_diffs>",
		"<git_log>",
		"<instruction>",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("XML missing %q", want)
		}
	}
	if !strings.HasSuffix(r.FileExtension(), ".xml") {
		t.Error("ext")
	}
}

func TestMarkdownRender(t *testing.T) {
	doc := sampleDoc(config.StyleMarkdown)
	r := For(config.StyleMarkdown)
	out, _ := r.Render(doc)
	for _, want := range []string{
		"# GopherCram Output",
		"## Directory Structure",
		"## Files",
		"### `src/main.go`",
		"```go",
		"## Git Log",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Markdown missing %q", want)
		}
	}
}

func TestJSONRender(t *testing.T) {
	doc := sampleDoc(config.StyleJSON)
	r := For(config.StyleJSON)
	out, err := r.Render(doc)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`"producer"`,
		`"files":`,
		`"src/main.go"`,
		`"git_log"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("JSON missing %q", want)
		}
	}
}

func TestPlainRender(t *testing.T) {
	doc := sampleDoc(config.StylePlain)
	r := For(config.StylePlain)
	out, _ := r.Render(doc)
	for _, want := range []string{
		"SUMMARY",
		"DIRECTORY STRUCTURE",
		"FILES",
		"---- src/main.go ",
		"GIT LOG",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Plain missing %q", want)
		}
	}
}

func TestRendererForUnknown(t *testing.T) {
	r := For("not-real")
	if r == nil {
		t.Fatal("expected fallback renderer")
	}
	if r.FileExtension() != ".xml" {
		t.Errorf("expected xml fallback, got %s", r.FileExtension())
	}
}

func TestSplit(t *testing.T) {
	doc := "line1\nline2\nline3\nline4\nline5\n"
	parts := Split(doc, 12)
	if len(parts) < 2 {
		t.Errorf("expected multiple parts, got %d", len(parts))
	}
	joined := strings.Join(parts, "\n")
	for _, want := range []string{"line1", "line5"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %q", want)
		}
	}
}

func TestSplit_Singleton(t *testing.T) {
	parts := Split("short", 100)
	if len(parts) != 1 {
		t.Error("small content should not split")
	}
}

func TestPartFilename(t *testing.T) {
	if got := PartFilename("repo.xml", 2); got != "repo.2.xml" {
		t.Errorf("got %s", got)
	}
	if got := PartFilename("a/b/c.md", 1); got != "a/b/c.1.md" {
		t.Errorf("got %s", got)
	}
}

func TestWriteCDATA_Escape(t *testing.T) {
	var b strings.Builder
	writeCDATA(&b, "before ]]> after")
	out := b.String()
	if !strings.Contains(out, "]]]]><![CDATA[>") {
		t.Errorf("CDATA terminator not split: %q", out)
	}
}

func TestXMLAttrEscape(t *testing.T) {
	got := xmlAttr(`a"b<c>&d`)
	if strings.ContainsAny(got, `"<>`) {
		t.Errorf("attr escape leaked: %s", got)
	}
}

func TestPickFenceLong(t *testing.T) {
	content := "code\n```\nmid\n````\nend\n"
	f := pickFence(content)
	if len(f) <= 4 {
		t.Errorf("expected longer fence, got %q", f)
	}
}

func TestLanguageFromPath(t *testing.T) {
	cases := map[string]string{
		"x.go":  "go",
		"x.py":  "python",
		"x.ts":  "typescript",
		"x.tsx": "tsx",
		"x.md":  "markdown",
		"x.zzz": "",
	}
	for in, want := range cases {
		if got := languageFromPath(in); got != want {
			t.Errorf("languageFromPath(%q)=%q, want %q", in, got, want)
		}
	}
}
