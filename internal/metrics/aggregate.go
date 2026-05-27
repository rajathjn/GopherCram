package metrics

import (
	"sort"
)

// FileMetrics holds per-file counts.
type FileMetrics struct {
	Path   string
	Chars  int
	Tokens int
	Lines  int
	Bytes  int64
}

// Aggregate captures the totals across the packed file set.
type Aggregate struct {
	TotalFiles  int
	TotalChars  int
	TotalTokens int
	TotalBytes  int64
	Files       []FileMetrics
}

// Top returns the n entries from the aggregate ordered by Tokens descending.
func (a Aggregate) Top(n int) []FileMetrics {
	if n <= 0 || len(a.Files) == 0 {
		return nil
	}
	sorted := make([]FileMetrics, len(a.Files))
	copy(sorted, a.Files)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Tokens != sorted[j].Tokens {
			return sorted[i].Tokens > sorted[j].Tokens
		}
		return sorted[i].Chars > sorted[j].Chars
	})
	if n > len(sorted) {
		n = len(sorted)
	}
	return sorted[:n]
}

// Compute calculates the aggregate metrics for a set of (path, content) pairs.
func Compute(items []struct {
	Path    string
	Content string
	Bytes   int64
}) Aggregate {
	agg := Aggregate{}
	for _, item := range items {
		fm := FileMetrics{
			Path:   item.Path,
			Chars:  CountRunes(item.Content),
			Tokens: EstimateTokens(item.Content),
			Lines:  CountLines(item.Content),
			Bytes:  item.Bytes,
		}
		agg.Files = append(agg.Files, fm)
		agg.TotalChars += fm.Chars
		agg.TotalTokens += fm.Tokens
		agg.TotalBytes += fm.Bytes
		agg.TotalFiles++
	}
	return agg
}
