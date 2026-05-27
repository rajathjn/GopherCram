package manipulate

import (
	"strings"
	"testing"

	"github.com/rajathjn/GopherCram/internal/config"
)

func TestApply_NoOp(t *testing.T) {
	in := "hello\nworld\n"
	out := Apply("x.go", in, config.Output{})
	if out != in {
		t.Errorf("no-op transform should pass through; got %q", out)
	}
}

func TestApply_TruncateBase64(t *testing.T) {
	long := strings.Repeat("A", 80)
	out := Apply("x.txt", long, config.Output{TruncateBase64: true})
	if strings.Contains(out, long) {
		t.Error("base64 not truncated")
	}
}

func TestApply_Compress(t *testing.T) {
	in := "func a() {\n  body()\n}\n"
	out := Apply("x.go", in, config.Output{Compress: true})
	if strings.Contains(out, "body()") {
		t.Errorf("body should be elided: %q", out)
	}
}

func TestApply_RemoveCommentsAndEmptyLines(t *testing.T) {
	in := "// header\n\npackage x\n\nfunc y() {}\n"
	out := Apply("x.go", in, config.Output{RemoveComments: true, RemoveEmptyLines: true})
	if strings.Contains(out, "header") {
		t.Error("comment leaked")
	}
	if strings.Contains(out, "\n\n") {
		t.Errorf("blank line leaked: %q", out)
	}
}

func TestApply_ShowLineNumbersWithRemove(t *testing.T) {
	in := "// gone\npackage x\nfunc y(){}\n"
	out := Apply("x.go", in, config.Output{RemoveComments: true, ShowLineNumbers: true})
	if !strings.Contains(out, " | ") {
		t.Errorf("missing line numbers: %q", out)
	}
}
