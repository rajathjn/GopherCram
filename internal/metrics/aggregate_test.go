package metrics

import "testing"

func TestCompute(t *testing.T) {
	items := []struct {
		Path    string
		Content string
		Bytes   int64
	}{
		{Path: "a", Content: "hello", Bytes: 5},
		{Path: "b", Content: "the quick brown fox", Bytes: 19},
	}
	agg := Compute(items)
	if agg.TotalFiles != 2 {
		t.Errorf("files=%d", agg.TotalFiles)
	}
	if agg.TotalBytes != 24 {
		t.Errorf("bytes=%d", agg.TotalBytes)
	}
	if agg.TotalChars != 24 {
		t.Errorf("chars=%d", agg.TotalChars)
	}
	if agg.TotalTokens <= 0 {
		t.Error("tokens should be positive")
	}
}

func TestAggregateTop(t *testing.T) {
	agg := Aggregate{
		Files: []FileMetrics{
			{Path: "a", Tokens: 5, Chars: 5},
			{Path: "b", Tokens: 50, Chars: 50},
			{Path: "c", Tokens: 25, Chars: 25},
		},
	}
	if got := agg.Top(0); got != nil {
		t.Error("Top(0) should be nil")
	}
	top := agg.Top(2)
	if len(top) != 2 || top[0].Path != "b" || top[1].Path != "c" {
		t.Errorf("unexpected top: %+v", top)
	}
	// More than length
	if got := agg.Top(10); len(got) != 3 {
		t.Errorf("expected 3, got %d", len(got))
	}
}

func TestAggregateTopEmpty(t *testing.T) {
	agg := Aggregate{}
	if got := agg.Top(5); got != nil {
		t.Error("expected nil from empty aggregate")
	}
}
