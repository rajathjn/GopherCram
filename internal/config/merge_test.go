package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromFile_BadJSON(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(p, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadFromFile(p); err == nil {
		t.Error("expected parse error")
	}
}

func TestLoadFromFile_Missing(t *testing.T) {
	cfg, err := LoadFromFile(filepath.Join(t.TempDir(), "nope.json"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg != nil {
		t.Error("expected nil for missing file")
	}
}

func TestLoadFromFile_Valid(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "c.json")
	if err := os.WriteFile(p, []byte(`{"output":{"style":"plain"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadFromFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Output.Style != StylePlain {
		t.Errorf("style=%v", cfg.Output.Style)
	}
}

func TestMerge_AllFields(t *testing.T) {
	base := Defaults()

	str := func(s string) *string { return &s }
	b := func(v bool) *bool { return &v }
	i := func(v int) *int { return &v }
	i64 := func(v int64) *int64 { return &v }
	style := StyleJSON

	p := PartialConfig{
		Schema: str("https://example/x.json"),
		Input:  &PartialInput{MaxFileSize: i64(1024)},
		Output: &PartialOutput{
			FilePath: str("o.json"), Style: &style,
			ParsableStyle: b(true), HeaderText: str("H"),
			InstructionFilePath: str("I"), FileSummary: b(false),
			DirectoryStructure: b(false), Files: b(false),
			RemoveComments: b(true), RemoveEmptyLines: b(true),
			Compress: b(true), TopFilesLength: i(11),
			ShowLineNumbers: b(true), TruncateBase64: b(true),
			CopyToClipboard: b(true), IncludeEmptyDirectories: b(true),
			IncludeFullDirectoryStructure: b(true),
			SplitOutputBytes:              i64(2048),
			Git: &PartialGit{
				SortByChanges: b(false), SortByChangesMaxCommits: i(50),
				IncludeDiffs: b(true), IncludeLogs: b(true), IncludeLogsCount: i(7),
			},
		},
		Include:    &[]string{"*.go"},
		Ignore:     &PartialIgnore{UseGitignore: b(false), UseDotIgnore: b(false), UseDefaultPatterns: b(false), CustomPatterns: &[]string{"x"}},
		Security:   &PartialSecurity{EnableSecurityCheck: b(false)},
		TokenCount: &PartialTokenCount{Encoding: str("approx2")},
	}
	m := Merge(base, p)
	if m.Schema != "https://example/x.json" {
		t.Errorf("schema=%q", m.Schema)
	}
	if m.Output.Style != StyleJSON || m.Output.FilePath != "o.json" {
		t.Errorf("output=%+v", m.Output)
	}
	if m.Output.Git.IncludeLogsCount != 7 {
		t.Errorf("git=%+v", m.Output.Git)
	}
	if m.Input.MaxFileSize != 1024 {
		t.Errorf("input=%+v", m.Input)
	}
	if m.Ignore.CustomPatterns[0] != "x" {
		t.Errorf("ignore=%+v", m.Ignore)
	}
	if m.Security.EnableSecurityCheck {
		t.Error("security off")
	}
	if m.TokenCount.Encoding != "approx2" {
		t.Errorf("encoding=%v", m.TokenCount.Encoding)
	}
}

func TestGlobalConfigPath_FallbackHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	if got := GlobalConfigPath(); got == "" {
		t.Skip("could not derive home dir")
	}
}
