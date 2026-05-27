package config

import (
	"strings"
	"testing"
)

func TestOutputStyleValid(t *testing.T) {
	for _, s := range []OutputStyle{StyleXML, StyleMarkdown, StyleJSON, StylePlain} {
		if !s.Valid() {
			t.Errorf("%s should be valid", s)
		}
	}
	if OutputStyle("nope").Valid() {
		t.Error("'nope' should be invalid")
	}
}

func TestDefaultFilenameFor(t *testing.T) {
	cases := map[OutputStyle]string{
		StyleXML:      "gophercram-output.xml",
		StyleMarkdown: "gophercram-output.md",
		StyleJSON:     "gophercram-output.json",
		StylePlain:    "gophercram-output.txt",
		"weird":       "gophercram-output.xml",
	}
	for in, want := range cases {
		if got := DefaultFilenameFor(in); got != want {
			t.Errorf("DefaultFilenameFor(%q)=%q, want %q", in, got, want)
		}
	}
}

func TestAllStyles(t *testing.T) {
	if got := len(AllStyles()); got != 4 {
		t.Errorf("AllStyles returned %d styles, want 4", got)
	}
}

func TestDefaults(t *testing.T) {
	d := Defaults()
	if d.Output.Style != StyleXML {
		t.Errorf("default style %q, want xml", d.Output.Style)
	}
	if !d.Ignore.UseGitignore {
		t.Error("UseGitignore default should be true")
	}
	if !d.Security.EnableSecurityCheck {
		t.Error("EnableSecurityCheck default should be true")
	}
	if d.Input.MaxFileSize == 0 {
		t.Error("MaxFileSize default should be non-zero")
	}
	if d.Output.TopFilesLength != 5 {
		t.Errorf("TopFilesLength default %d, want 5", d.Output.TopFilesLength)
	}
}

func TestValidate(t *testing.T) {
	d := Defaults()
	if err := d.Validate(); err != nil {
		t.Fatalf("default config should validate: %v", err)
	}
	bad := d
	bad.Output.Style = "nope"
	if err := bad.Validate(); err == nil {
		t.Error("expected invalid style to fail validation")
	}

	for name, mutate := range map[string]func(*Config){
		"max-file-size": func(c *Config) { c.Input.MaxFileSize = -1 },
		"top-files":     func(c *Config) { c.Output.TopFilesLength = -1 },
		"logs-count":    func(c *Config) { c.Output.Git.IncludeLogsCount = -1 },
		"split-output":  func(c *Config) { c.Output.SplitOutputBytes = -1 },
		"include-blank": func(c *Config) { c.Include = []string{"   "} },
	} {
		t.Run(name, func(t *testing.T) {
			c := Defaults()
			mutate(&c)
			if err := c.Validate(); err == nil {
				t.Error("expected validation error")
			}
		})
	}
}

func TestDefaultIgnorePatterns(t *testing.T) {
	pats := DefaultIgnorePatterns()
	if len(pats) < 10 {
		t.Errorf("expected many default patterns, got %d", len(pats))
	}
	found := 0
	for _, p := range pats {
		if strings.Contains(p, "node_modules") || strings.Contains(p, ".git") {
			found++
		}
	}
	if found < 2 {
		t.Error("default patterns should ignore node_modules and .git")
	}
}

func TestBinaryFileExtensions(t *testing.T) {
	exts := BinaryFileExtensions()
	for _, want := range []string{".png", ".zip", ".pdf", ".exe"} {
		if _, ok := exts[want]; !ok {
			t.Errorf("expected %s in binary extensions", want)
		}
	}
}
