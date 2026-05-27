package fileselect

import (
	"bytes"
	"errors"
	"os"
	"unicode/utf8"
)

// File is a file result paired with its in-memory content.
type File struct {
	Result
	Content string
	// Skipped is set when the file was filtered out after read (e.g. binary
	// sniff detected a non-text file).
	Skipped bool
	// SkipReason gives a short human-readable explanation.
	SkipReason string
}

// ReadAll loads the content of every selected file. Binary files (by extension
// or by content sniff) are recorded but their content is stored as the empty
// string with Skipped=true.
func ReadAll(selected []Result) ([]File, error) {
	out := make([]File, 0, len(selected))
	for _, r := range selected {
		f := File{Result: r}
		if r.IsBinary {
			f.Skipped = true
			f.SkipReason = "binary file (by extension)"
			out = append(out, f)
			continue
		}

		data, err := os.ReadFile(r.AbsPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			f.Skipped = true
			f.SkipReason = "read error: " + err.Error()
			out = append(out, f)
			continue
		}

		if looksBinary(data) {
			f.Skipped = true
			f.SkipReason = "binary file (by content)"
			f.IsBinary = true
			out = append(out, f)
			continue
		}

		f.Content = string(data)
		out = append(out, f)
	}
	return out, nil
}

// looksBinary reports whether a byte slice looks like binary content. We use
// the same heuristic as `git` and `grep`: presence of a NUL byte in the first
// 8 KiB, plus a check that the content is mostly valid UTF-8.
func looksBinary(data []byte) bool {
	const sniff = 8 * 1024
	if len(data) > sniff {
		data = data[:sniff]
	}
	if bytes.IndexByte(data, 0) >= 0 {
		return true
	}
	if !utf8.Valid(data) {
		// Count invalid runes; if more than 10% of bytes are invalid, treat
		// the data as binary.
		invalid := 0
		i := 0
		for i < len(data) {
			r, size := utf8.DecodeRune(data[i:])
			if r == utf8.RuneError && size == 1 {
				invalid++
			}
			i += size
		}
		if len(data) > 0 && float64(invalid)/float64(len(data)) > 0.10 {
			return true
		}
	}
	return false
}
