package output

import (
	"path/filepath"
	"strconv"
	"strings"
)

// Split breaks `content` into multiple chunks each at most `maxBytes` long.
// We try to break at line boundaries to avoid splitting in the middle of a
// file's contents.
func Split(content string, maxBytes int64) []string {
	if maxBytes <= 0 || int64(len(content)) <= maxBytes {
		return []string{content}
	}
	var chunks []string
	remaining := content
	for int64(len(remaining)) > maxBytes {
		// find last newline within the window
		cut := int(maxBytes)
		idx := strings.LastIndexByte(remaining[:cut], '\n')
		if idx <= 0 {
			idx = cut
		}
		chunks = append(chunks, remaining[:idx])
		remaining = remaining[idx:]
		remaining = strings.TrimPrefix(remaining, "\n")
	}
	if remaining != "" {
		chunks = append(chunks, remaining)
	}
	return chunks
}

// PartFilename returns the i-th part filename (1-based) for a given output
// path, e.g. ("repo.xml", 2) → "repo.2.xml".
func PartFilename(base string, idx int) string {
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	return stem + "." + strconv.Itoa(idx) + ext
}
