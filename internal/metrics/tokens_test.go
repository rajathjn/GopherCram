package metrics

import "testing"

func TestEstimateTokens_Empty(t *testing.T) {
	if got := EstimateTokens(""); got != 0 {
		t.Errorf("expected 0 for empty, got %d", got)
	}
}

func TestEstimateTokens_Plain(t *testing.T) {
	n := EstimateTokens("hello world this is a sentence")
	if n < 4 || n > 12 {
		t.Errorf("token count out of range: %d", n)
	}
}

func TestEstimateTokens_Code(t *testing.T) {
	n := EstimateTokens(`func main() { fmt.Println("hi") }`)
	if n < 6 {
		t.Errorf("code token count too low: %d", n)
	}
}

func TestEstimateTokens_Unicode(t *testing.T) {
	n := EstimateTokens("こんにちは世界")
	if n < 5 {
		t.Errorf("unicode count too low: %d", n)
	}
}

func TestCountRunes(t *testing.T) {
	if CountRunes("héllo") != 5 {
		t.Error("expected 5 runes")
	}
}

func TestCountLines(t *testing.T) {
	cases := map[string]int{
		"":          0,
		"a":         1,
		"a\nb":      2,
		"a\nb\n":    2,
		"a\nb\nc\n": 3,
	}
	for in, want := range cases {
		if got := CountLines(in); got != want {
			t.Errorf("CountLines(%q)=%d, want %d", in, got, want)
		}
	}
}

func TestEstimateUnicodeTokens_Empty(t *testing.T) {
	if got := estimateUnicodeTokens(""); got != 1 {
		t.Errorf("got %d", got)
	}
}

func TestEstimateASCIIDigitsAndCase(t *testing.T) {
	n := EstimateTokens("camelCaseIdentifier and SHOUTING and snake_case_long and 1234567890")
	if n < 5 {
		t.Errorf("got %d", n)
	}
}
