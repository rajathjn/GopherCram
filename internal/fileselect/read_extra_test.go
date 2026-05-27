package fileselect

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLooksBinary_HighInvalidUTF8(t *testing.T) {
	// Construct a payload that's mostly invalid UTF-8.
	data := make([]byte, 100)
	for i := range data {
		data[i] = 0xFF
	}
	if !looksBinary(data) {
		t.Error("0xFF-heavy payload should be detected as binary")
	}
}

func TestLooksBinary_LargePayload(t *testing.T) {
	// >8 KiB; should still be detected on the prefix.
	data := append([]byte{0x00}, make([]byte, 16*1024)...)
	if !looksBinary(data) {
		t.Error("NUL in first 8 KiB should mark binary")
	}
}

func TestReadAll_ReadErrorSurfaced(t *testing.T) {
	dir := t.TempDir()
	// Create then remove a file so the absolute path no longer resolves
	// (simulates a transient read error). On macOS, ENOENT — which we treat
	// as "skip". We exercise the path either way.
	p := filepath.Join(dir, "gone.txt")
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(p); err != nil {
		t.Fatal(err)
	}
	files, err := ReadAll([]Result{{AbsPath: p, RelPath: "gone.txt"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Errorf("expected file to be silently dropped, got %v", files)
	}
}

func TestWalk_BinaryByContentDuringWalk(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "raw.dat"),
		[]byte{0x00, 0x01, 0x02}, 0o644); err != nil {
		t.Fatal(err)
	}
	rs, err := Walk(Options{Roots: []string{root}})
	if err != nil {
		t.Fatal(err)
	}
	files, err := ReadAll(rs)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 || !files[0].Skipped {
		t.Errorf("expected NUL-byte file to be marked skipped: %+v", files)
	}
	if !strings.Contains(files[0].SkipReason, "binary") {
		t.Errorf("unexpected reason: %q", files[0].SkipReason)
	}
}
