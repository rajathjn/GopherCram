package fileselect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadAll_Text(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "a.txt")
	if err := os.WriteFile(file, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	res := []Result{{AbsPath: file, RelPath: "a.txt", Size: 5}}
	files, err := ReadAll(res)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].Content != "hello" {
		t.Errorf("unexpected: %+v", files)
	}
}

func TestReadAll_BinaryByExt(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "x.png")
	if err := os.WriteFile(file, []byte{0x89, 0x50, 0x4E, 0x47}, 0o644); err != nil {
		t.Fatal(err)
	}
	files, _ := ReadAll([]Result{{AbsPath: file, RelPath: "x.png", IsBinary: true}})
	if !files[0].Skipped {
		t.Error("expected binary file to be skipped")
	}
}

func TestReadAll_BinaryByContent(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "x.dat")
	if err := os.WriteFile(file, []byte{0x00, 0x01, 0x02, 0x03}, 0o644); err != nil {
		t.Fatal(err)
	}
	files, _ := ReadAll([]Result{{AbsPath: file, RelPath: "x.dat"}})
	if !files[0].Skipped {
		t.Error("expected NUL-byte file to be detected as binary")
	}
}

func TestReadAll_MissingFile(t *testing.T) {
	files, err := ReadAll([]Result{{AbsPath: "/nope/missing", RelPath: "missing"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

func TestLooksBinary(t *testing.T) {
	if !looksBinary([]byte{0x00, 'a', 'b'}) {
		t.Error("NUL byte should detect binary")
	}
	if looksBinary([]byte("plain text\nhello")) {
		t.Error("text should not be binary")
	}
}
