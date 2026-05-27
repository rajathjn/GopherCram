package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPartialFromFile_BadJSON(t *testing.T) {
	p := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(p, []byte("{garbage"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadPartialFromFile(p); err == nil {
		t.Error("expected error parsing bad JSON")
	}
}

func TestStripJSONCommentsEscapesInString(t *testing.T) {
	// Escaped quote inside a string should not terminate the string and
	// confuse the comment stripper.
	in := []byte(`{"a": "x\"y // not a comment", "b": 1 // gone}`)
	out := stripJSONComments(in).String()
	if contains(out, "gone") {
		t.Errorf("line comment leaked: %q", out)
	}
	if !contains(out, `x\"y // not a comment`) {
		t.Errorf("string content broken: %q", out)
	}
}
