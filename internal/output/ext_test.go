package output

import "testing"

func TestFileExtensions(t *testing.T) {
	cases := map[string]Renderer{
		".xml": XMLRenderer{},
		".md":  MarkdownRenderer{},
		".json": JSONRenderer{},
		".txt": PlainRenderer{},
	}
	for want, r := range cases {
		if got := r.FileExtension(); got != want {
			t.Errorf("%T.FileExtension()=%q, want %q", r, got, want)
		}
	}
}
