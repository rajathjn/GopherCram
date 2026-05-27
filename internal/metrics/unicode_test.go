package metrics

import "testing"

func TestEstimateUnicodeTokens_Punctuation(t *testing.T) {
	// Mixed punctuation, whitespace, and CJK content.
	n := estimateUnicodeTokens("こんにちは、世界！\n夢を見る。")
	if n < 6 {
		t.Errorf("expected >=6 tokens, got %d", n)
	}
}

func TestEstimateUnicodeTokens_WhitespaceOnly(t *testing.T) {
	if got := estimateUnicodeTokens("   \t  "); got != 1 {
		t.Errorf("expected 1 token for whitespace-only, got %d", got)
	}
}

func TestEstimateTokens_HighUnicode(t *testing.T) {
	// emoji-heavy content should route to estimateUnicodeTokens
	n := EstimateTokens("🚀🚀🚀 hello 世界 🎉")
	if n < 4 {
		t.Errorf("got %d tokens", n)
	}
}
