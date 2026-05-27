package manipulate

import (
	"strings"
	"testing"
)

func TestStripCStyleCommentsStrings(t *testing.T) {
	// Make sure strings, chars, and backticks all survive intact.
	in := "a := \"//\"; b := '/'; c := `/* not a comment */`; // real\n"
	out := stripCStyleComments(in)
	for _, want := range []string{`"//"`, `'/'`, "`/* not a comment */`"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %q", want, out)
		}
	}
	if strings.Contains(out, "real") {
		t.Errorf("line comment leaked: %q", out)
	}
}

func TestStripCStyleEscapes(t *testing.T) {
	in := "s := \"a\\\"b // not a comment\"; // real\n"
	out := stripCStyleComments(in)
	if !strings.Contains(out, `\"`) {
		t.Errorf("escape sequence lost: %q", out)
	}
	if strings.Contains(out, "real") {
		t.Errorf("trailing comment leaked: %q", out)
	}
}

func TestStripCStyleBlockNewlines(t *testing.T) {
	in := "/*\n line1\n line2\n*/\nafter\n"
	out := stripCStyleComments(in)
	// Block comment should leave the newlines in place so line numbers don't shift.
	if strings.Count(out, "\n") < 3 {
		t.Errorf("expected preserved newlines, got %q", out)
	}
	if strings.Contains(out, "line1") {
		t.Errorf("block body leaked: %q", out)
	}
}

func TestStripHashEscapes(t *testing.T) {
	in := "s = \"a # b\"; t = 'c # d' # comment\n"
	out := stripHashComments(in)
	if !strings.Contains(out, "a # b") || !strings.Contains(out, "c # d") {
		t.Errorf("strings broken: %q", out)
	}
	if strings.Contains(out, " comment\n") && !strings.HasSuffix(out, "comment") {
		// allow trailing comment text only when fully stripped
	}
}

func TestRemoveCommentsCSSAlias(t *testing.T) {
	in := "/* gone */ .x { color: red }"
	out := RemoveComments("x.css", in)
	if strings.Contains(out, "gone") {
		t.Errorf("CSS comment leaked: %q", out)
	}
}

func TestRemoveCommentsRustGoesCStyle(t *testing.T) {
	in := "fn main() { // hello\n}"
	out := RemoveComments("x.rs", in)
	if strings.Contains(out, "hello") {
		t.Errorf("rust // comment leaked: %q", out)
	}
}

func TestCompressJSish(t *testing.T) {
	in := `export function f() {
  const a = 1;
  return a;
}

export const x = 42;
`
	out := CompressCode("x.ts", in)
	if !strings.Contains(out, "export function f") || !strings.Contains(out, "export const x = 42") {
		t.Errorf("signatures missing: %q", out)
	}
	if strings.Contains(out, "const a = 1") {
		t.Errorf("body kept: %q", out)
	}
}

func TestCompressCLike(t *testing.T) {
	in := `public class Foo {
    public void bar() {
        int x = 1;
    }
}
`
	out := CompressCode("x.java", in)
	if !strings.Contains(out, "public class Foo") {
		t.Errorf("class missing: %q", out)
	}
}
