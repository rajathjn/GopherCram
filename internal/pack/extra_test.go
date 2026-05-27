package pack

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rajathjn/GopherCram/internal/config"
)

func TestPack_FullDirectoryStructure(t *testing.T) {
	dir := makeFixture(t)
	cfg := config.Defaults()
	cfg.Cwd = dir
	cfg.Include = []string{"src/api/*.go"} // narrow include, but ask for full tree
	cfg.Output.IncludeFullDirectoryStructure = true

	pk := New("test", "0.0.1")
	res, err := pk.Pack(&cfg, []string{dir})
	if err != nil {
		t.Fatal(err)
	}
	// The included file set is narrowed by the include filter, but the tree
	// section should still mention paths outside of `src/api/`.
	if !strings.Contains(res.Output, "README.md") {
		t.Errorf("full directory tree should include README.md: %s", res.Output[:200])
	}
}

func TestPack_Instruction(t *testing.T) {
	dir := makeFixture(t)
	instr := filepath.Join(dir, "instruction.txt")
	if err := os.WriteFile(instr, []byte("focus on the API layer"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Defaults()
	cfg.Cwd = dir
	cfg.Output.InstructionFilePath = instr
	pk := New("test", "0.0.1")
	res, err := pk.Pack(&cfg, []string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.Output, "focus on the API layer") {
		t.Error("instruction text missing")
	}
}

func TestPack_HeaderText(t *testing.T) {
	dir := makeFixture(t)
	cfg := config.Defaults()
	cfg.Cwd = dir
	cfg.Output.HeaderText = "review this code"
	pk := New("test", "0.0.1")
	res, err := pk.Pack(&cfg, []string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.Output, "review this code") {
		t.Error("header text missing")
	}
}

func TestWriteToDisk_AbsolutePath(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Defaults()
	cfg.Cwd = "/tmp" // intentionally different
	abs := filepath.Join(dir, "deep", "nested", "out.xml")
	cfg.Output.FilePath = abs
	res := &Result{Output: "<x/>"}
	written, err := WriteToDisk(&cfg, res)
	if err != nil {
		t.Fatal(err)
	}
	if written[0] != abs {
		t.Errorf("expected %s, got %s", abs, written[0])
	}
}
