package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPartialFromFile_Missing(t *testing.T) {
	p, err := LoadPartialFromFile(filepath.Join(t.TempDir(), "nope.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p != nil {
		t.Error("expected nil partial for missing file")
	}
}

func TestLoadPartialFromFile_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.json")
	body := `{
        "output": {
            "style": "markdown",
            "fileSummary": false,
            "topFilesLength": 12,
            "git": {"includeDiffs": true}
        },
        "ignore": {"useGitignore": false, "customPatterns": ["*.tmp"]},
        "security": {"enableSecurityCheck": false},
        "include": ["src/**/*.ts"]
    }`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	p, err := LoadPartialFromFile(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if p == nil {
		t.Fatal("expected partial")
	}
	merged := Merge(Defaults(), *p)
	if merged.Output.Style != StyleMarkdown {
		t.Errorf("style=%v", merged.Output.Style)
	}
	if merged.Output.FileSummary {
		t.Error("fileSummary should be false after merge")
	}
	if merged.Output.TopFilesLength != 12 {
		t.Errorf("topFilesLength=%d", merged.Output.TopFilesLength)
	}
	if !merged.Output.Git.IncludeDiffs {
		t.Error("IncludeDiffs should be true")
	}
	if merged.Ignore.UseGitignore {
		t.Error("UseGitignore should be false")
	}
	if len(merged.Ignore.CustomPatterns) != 1 || merged.Ignore.CustomPatterns[0] != "*.tmp" {
		t.Errorf("CustomPatterns=%v", merged.Ignore.CustomPatterns)
	}
	if merged.Security.EnableSecurityCheck {
		t.Error("security check should be off")
	}
	if len(merged.Include) != 1 {
		t.Errorf("include patterns=%v", merged.Include)
	}
}

func TestLoadPartialFromFile_WithComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.jsonc")
	body := `{
        // this is a comment
        /* block too */
        "output": {"style": "json"}
    }`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	// LoadPartialFromFile uses strict JSON; only LoadFromFile strips comments.
	cfg, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected cfg")
	}
	if cfg.Output.Style != StyleJSON {
		t.Errorf("style=%v", cfg.Output.Style)
	}
}

func TestDiscover(t *testing.T) {
	dir := t.TempDir()
	if got := Discover(dir); got != "" {
		t.Errorf("expected empty discover, got %q", got)
	}
	cfgPath := filepath.Join(dir, ConfigFilename)
	if err := os.WriteFile(cfgPath, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := Discover(dir); got != cfgPath {
		t.Errorf("Discover=%q, want %q", got, cfgPath)
	}
}

func TestDiscover_Alternate(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "repomix.config.json")
	if err := os.WriteFile(cfgPath, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := Discover(dir); got != cfgPath {
		t.Errorf("Discover=%q, want %q", got, cfgPath)
	}
}

func TestGlobalConfigPath(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/test")
	if got := GlobalConfigPath(); got == "" {
		t.Error("expected non-empty path")
	}
}

func TestMerge_BoolNegation(t *testing.T) {
	base := Defaults()
	if !base.Output.FileSummary {
		t.Fatal("baseline expectation broken")
	}
	off := false
	p := PartialConfig{Output: &PartialOutput{FileSummary: &off}}
	merged := Merge(base, p)
	if merged.Output.FileSummary {
		t.Error("FileSummary should now be false")
	}
}

func TestStripJSONComments(t *testing.T) {
	in := []byte(`{
        // line comment
        "a": "v // not a comment",
        /* block
           still in block */
        "b": 1
    }`)
	out := stripJSONComments(in).String()
	if !contains(out, `"a": "v // not a comment"`) {
		t.Error("string literal // got stripped")
	}
	if contains(out, "block") {
		t.Error("block comment leaked through")
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
