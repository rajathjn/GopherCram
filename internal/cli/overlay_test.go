package cli

import (
	"testing"

	"github.com/rajathjn/GopherCram/internal/config"
)

func TestOverlayFromArgsNil(t *testing.T) {
	p, err := OverlayFromArgs(nil)
	if err != nil {
		t.Fatal(err)
	}
	if p == nil {
		t.Error("expected non-nil overlay")
	}
}

func TestOverlayFromArgsFlags(t *testing.T) {
	parser := BuildParser()
	args, err := parser.Parse([]string{
		"--style", "markdown",
		"--output", "out.md",
		"--include", "src/**/*.ts,tests/**",
		"--ignore", "*.tmp",
		"--no-security-check",
		"--compress",
		"--max-file-size", "1024",
		"--split-output", "1mb",
		"--top-files-len", "7",
	})
	if err != nil {
		t.Fatal(err)
	}
	overlay, err := OverlayFromArgs(args)
	if err != nil {
		t.Fatal(err)
	}
	merged := config.Merge(config.Defaults(), *overlay)
	if merged.Output.Style != config.StyleMarkdown {
		t.Errorf("style=%v", merged.Output.Style)
	}
	if merged.Output.FilePath != "out.md" {
		t.Errorf("output=%v", merged.Output.FilePath)
	}
	if len(merged.Include) != 2 {
		t.Errorf("include=%v", merged.Include)
	}
	if len(merged.Ignore.CustomPatterns) != 1 || merged.Ignore.CustomPatterns[0] != "*.tmp" {
		t.Errorf("custom=%v", merged.Ignore.CustomPatterns)
	}
	if merged.Security.EnableSecurityCheck {
		t.Error("security should be off")
	}
	if !merged.Output.Compress {
		t.Error("compress should be on")
	}
	if merged.Input.MaxFileSize != 1024 {
		t.Errorf("maxsize=%d", merged.Input.MaxFileSize)
	}
	if merged.Output.SplitOutputBytes != 1024*1024 {
		t.Errorf("split=%d", merged.Output.SplitOutputBytes)
	}
	if merged.Output.TopFilesLength != 7 {
		t.Errorf("top=%d", merged.Output.TopFilesLength)
	}
}

func TestOverlayBadSplit(t *testing.T) {
	parser := BuildParser()
	args, err := parser.Parse([]string{"--split-output", "garbage"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := OverlayFromArgs(args); err == nil {
		t.Error("expected error parsing bad size")
	}
}

func TestSplitMultiCSV(t *testing.T) {
	got := splitMultiCSV([]string{"a,b", "c, d"})
	want := []string{"a", "b", "c", "d"}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i, v := range want {
		if got[i] != v {
			t.Errorf("got[%d]=%q, want %q", i, got[i], v)
		}
	}
}
