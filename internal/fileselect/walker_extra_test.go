package fileselect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWalk_NoRoots_DefaultsToDot(t *testing.T) {
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "x.go"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	res, err := Walk(Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res) == 0 {
		t.Error("expected at least one file when defaulting to '.'")
	}
}

func TestWalk_BadRoot(t *testing.T) {
	if _, err := Walk(Options{Roots: []string{"/definitely/missing/path/123"}}); err == nil {
		t.Error("expected error for missing root")
	}
}

func TestWalk_DedupAcrossRoots(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "x.go"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Walk the same directory twice — results should not be duplicated.
	res, err := Walk(Options{Roots: []string{dir, dir}})
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 1 {
		t.Errorf("expected deduplicated single result, got %d", len(res))
	}
}
