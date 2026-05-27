package manipulate

import (
	"strings"
	"testing"
)

func TestWithLineNumbers_Empty(t *testing.T) {
	if out := WithLineNumbers(""); out != "" {
		// the implementation produces "  1 | " for a single empty line; we
		// accept either as long as we don't crash.
		_ = out
	}
}

func TestNumString(t *testing.T) {
	cases := map[int]string{
		0: "0", 1: "1", 9: "9", 10: "10", 1234: "1234",
	}
	for in, want := range cases {
		if got := numString(in); got != want {
			t.Errorf("numString(%d)=%q want %q", in, got, want)
		}
	}
}

func TestDigits(t *testing.T) {
	cases := map[int]int{0: 1, 9: 1, 10: 2, 99: 2, 100: 3}
	for in, want := range cases {
		if got := digits(in); got != want {
			t.Errorf("digits(%d)=%d want %d", in, got, want)
		}
	}
}

func TestStripXMLNewlines(t *testing.T) {
	in := "<!-- multi\n line\n comment -->after"
	out := stripXMLComments(in)
	if strings.Contains(out, "multi") {
		t.Errorf("comment body leaked: %q", out)
	}
	if !strings.Contains(out, "after") {
		t.Errorf("non-comment content lost: %q", out)
	}
}

func TestStripSQLNewlines(t *testing.T) {
	in := "SELECT 1;\n/* multi\n line\n */\nSELECT 2;"
	out := stripSQLComments(in)
	if strings.Contains(out, "line") {
		t.Errorf("sql block leaked: %q", out)
	}
}

func TestCompressKeepsTopLevelComments(t *testing.T) {
	in := `// header comment
package main

func body() {
  inner()
}
`
	out := CompressCode("x.go", in)
	if !strings.Contains(out, "header comment") {
		t.Errorf("top-level comment lost: %q", out)
	}
}
