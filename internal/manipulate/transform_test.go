package manipulate

import (
	"strings"
	"testing"

	"github.com/rajathjn/GopherCram/internal/config"
)

func TestRemoveEmptyLines(t *testing.T) {
	got := RemoveEmptyLines("a\n\nb\n  \nc")
	if got != "a\nb\nc" {
		t.Errorf("got %q", got)
	}
}

func TestWithLineNumbers(t *testing.T) {
	out := WithLineNumbers("a\nb\nc")
	if !strings.Contains(out, "1 | a") || !strings.Contains(out, "3 | c") {
		t.Errorf("missing numbers: %q", out)
	}
}

func TestWithLineNumbers_TrailingNewline(t *testing.T) {
	out := WithLineNumbers("only\n")
	if !strings.HasSuffix(out, "only") {
		t.Errorf("unexpected: %q", out)
	}
}

func TestTruncateBase64(t *testing.T) {
	long := strings.Repeat("A", 80) + "=="
	in := "data: " + long + " more"
	out := TruncateBase64(in)
	if strings.Contains(out, long) {
		t.Error("expected base64 to be replaced")
	}
	if !strings.Contains(out, "[base64 omitted") {
		t.Error("placeholder missing")
	}
}

func TestRemoveCommentsCStyle(t *testing.T) {
	in := `// top
package main
/* block
   comment */
func main() { // inline
	s := "// not a comment"
}`
	out := RemoveComments("x.go", in)
	if strings.Contains(out, "top") || strings.Contains(out, "block") {
		t.Errorf("comments leaked: %q", out)
	}
	if !strings.Contains(out, `"// not a comment"`) {
		t.Error("string literal was stripped")
	}
}

func TestRemoveCommentsHash(t *testing.T) {
	in := "# greeting\nprint('hi') # trailing"
	out := RemoveComments("x.py", in)
	if strings.Contains(out, "greeting") {
		t.Error("python comment leaked")
	}
}

func TestRemoveCommentsXML(t *testing.T) {
	in := "<a>1</a><!-- c --><b>2</b>"
	out := RemoveComments("x.html", in)
	if strings.Contains(out, "c ") {
		t.Errorf("html comment leaked: %q", out)
	}
}

func TestRemoveCommentsLua(t *testing.T) {
	in := "x = 1 -- comment\n--[[ block\nblock ]]\ny=2"
	out := RemoveComments("x.lua", in)
	if strings.Contains(out, "comment") || strings.Contains(out, "block") {
		t.Errorf("lua comments leaked: %q", out)
	}
}

func TestRemoveCommentsSQL(t *testing.T) {
	in := "SELECT 1 -- comment\n/* block */"
	out := RemoveComments("x.sql", in)
	if strings.Contains(out, "comment") || strings.Contains(out, "block") {
		t.Errorf("sql comments leaked: %q", out)
	}
}

func TestApplyOrder(t *testing.T) {
	opts := config.Output{
		RemoveComments:   true,
		RemoveEmptyLines: true,
		ShowLineNumbers:  true,
	}
	in := "// top\npackage main\n\nfunc x() {}\n"
	out := Apply("x.go", in, opts)
	if strings.Contains(out, "top") {
		t.Error("comments not removed")
	}
	if !strings.Contains(out, " | ") {
		t.Error("line numbers not applied")
	}
}

func TestCompressGo(t *testing.T) {
	in := `package main

func long() int {
	a := 1
	b := 2
	return a + b
}

type Foo struct {
	A int
}
`
	out := CompressCode("x.go", in)
	if !strings.Contains(out, "func long()") {
		t.Errorf("compress dropped signature: %q", out)
	}
	if strings.Contains(out, "a := 1") {
		t.Errorf("compress kept body: %q", out)
	}
}

func TestCompressPython(t *testing.T) {
	in := `def hello():
    print("hi")
    print("there")

class C:
    def m(self):
        x = 1
        return x
`
	out := CompressCode("x.py", in)
	if !strings.Contains(out, "def hello") || !strings.Contains(out, "class C") {
		t.Errorf("compress python signatures missing: %q", out)
	}
	if strings.Contains(out, "x = 1") {
		t.Errorf("python body kept: %q", out)
	}
}

func TestCompressUnknownExtPassThrough(t *testing.T) {
	in := "anything\nat all"
	if CompressCode("x.unknownext", in) != in {
		t.Error("unknown ext should pass through")
	}
}
