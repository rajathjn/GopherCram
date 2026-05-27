package ignore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPatternSetBasic(t *testing.T) {
	ps := New()
	ps.AddAll([]string{
		"*.log",
		"build/",
		"!build/keep.txt",
		"# a comment",
		"",
	})
	if ps.Len() == 0 {
		t.Fatal("expected rules")
	}
	if !ps.MatchesPath("x.log", false) {
		t.Error("*.log should match")
	}
	if !ps.MatchesPath("build", true) {
		t.Error("build/ should ignore build dir")
	}
	if ps.MatchesPath("build/keep.txt", false) {
		t.Error("negation should re-include build/keep.txt")
	}
}

func TestPatternSetLoadFile(t *testing.T) {
	dir := t.TempDir()
	gi := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(gi, []byte("*.tmp\n!important.tmp\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ps := New()
	if err := ps.LoadFile(gi); err != nil {
		t.Fatal(err)
	}
	if !ps.MatchesPath("foo.tmp", false) {
		t.Error("foo.tmp should match")
	}
	if ps.MatchesPath("important.tmp", false) {
		t.Error("important.tmp should be re-included")
	}
}

func TestPatternSetLoadMissing(t *testing.T) {
	ps := New()
	if err := ps.LoadFile(filepath.Join(t.TempDir(), "nope")); err != nil {
		t.Error(err)
	}
}

func TestPatternSetLoadReader(t *testing.T) {
	ps := New()
	if err := ps.LoadReader(strings.NewReader("a.txt\nb.txt\n")); err != nil {
		t.Fatal(err)
	}
	if ps.Len() != 2 {
		t.Fatalf("len=%d", ps.Len())
	}
}

func TestPatternSetMerge(t *testing.T) {
	a := New()
	a.Add("*.log")
	b := New()
	b.Add("*.tmp")
	a.Merge(b)
	if a.Len() != 2 {
		t.Errorf("len=%d", a.Len())
	}
	a.Merge(nil)
}

func TestPatternSetNilSafe(t *testing.T) {
	var ps *PatternSet
	if ps.MatchesPath("x", false) {
		t.Error("nil pattern set should match nothing")
	}
}

func TestAnchoredVsFloating(t *testing.T) {
	ps := New()
	ps.Add("/main.go")
	if !ps.MatchesPath("main.go", false) {
		t.Error("anchored /main.go should match top-level")
	}
	if ps.MatchesPath("pkg/main.go", false) {
		t.Error("anchored /main.go should not match nested")
	}
}
