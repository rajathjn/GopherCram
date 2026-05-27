package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rajathjn/GopherCram/internal/metrics"
	"github.com/rajathjn/GopherCram/internal/pack"
)

func TestWriteTokenTree(t *testing.T) {
	res := &pack.Result{
		Metrics: metrics.Aggregate{
			TotalFiles: 2,
			Files: []metrics.FileMetrics{
				{Path: "a.go", Tokens: 50},
				{Path: "b.go", Tokens: 10},
			},
		},
	}
	var buf bytes.Buffer
	writeTokenTree(&buf, res, 0)
	out := buf.String()
	if !strings.Contains(out, "a.go") || !strings.Contains(out, "b.go") {
		t.Errorf("missing entries: %q", out)
	}

	buf.Reset()
	writeTokenTree(&buf, res, 30)
	out = buf.String()
	if !strings.Contains(out, "a.go") {
		t.Errorf("expected a.go above threshold: %q", out)
	}
	if strings.Contains(out, "b.go\t") {
		t.Errorf("did not expect b.go above threshold: %q", out)
	}
}

func TestWriteTokenTree_Empty(t *testing.T) {
	var buf bytes.Buffer
	writeTokenTree(&buf, &pack.Result{}, 0)
	if buf.Len() != 0 {
		t.Errorf("expected no output for empty aggregate, got %q", buf.String())
	}
}
