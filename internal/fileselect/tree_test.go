package fileselect

import (
	"strings"
	"testing"
)

func TestBuildTreeAndRender(t *testing.T) {
	paths := []string{
		"src/main.go",
		"src/util/helpers.go",
		"docs/readme.md",
		"LICENSE",
	}
	root := BuildTree(paths)
	out := RenderTree(root)
	for _, want := range []string{"src/", "util/", "main.go", "helpers.go", "docs/", "LICENSE"} {
		if !strings.Contains(out, want) {
			t.Errorf("rendered tree missing %q\n%s", want, out)
		}
	}
}

func TestBuildTree_Empty(t *testing.T) {
	root := BuildTree(nil)
	out := RenderTree(root)
	if out != "" {
		t.Errorf("expected empty render, got %q", out)
	}
}

func TestPathsOf(t *testing.T) {
	got := PathsOf([]Result{{RelPath: "a"}, {RelPath: "b"}})
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("PathsOf=%v", got)
	}
}
