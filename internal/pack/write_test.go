package pack

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rajathjn/GopherCram/internal/config"
)

func TestWriteToDisk_MkdirError(t *testing.T) {
	// Make the would-be parent directory be a regular file so MkdirAll fails.
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Defaults()
	cfg.Cwd = dir
	cfg.Output.FilePath = filepath.Join(blocker, "inner.xml")
	res := &Result{Output: "<x/>"}
	if _, err := WriteToDisk(&cfg, res); err == nil {
		t.Error("expected mkdir error")
	}
}

func TestWriteToDisk_DefaultFilenameWhenEmpty(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Defaults()
	cfg.Cwd = dir
	cfg.Output.FilePath = "" // force fallback
	res := &Result{Output: "<x/>"}
	written, err := WriteToDisk(&cfg, res)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(written[0]) != "gophercram-output.xml" {
		t.Errorf("expected default filename, got %s", written[0])
	}
}
