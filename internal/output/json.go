package output

import (
	"encoding/json"
)

// JSONRenderer emits a structured JSON document.
type JSONRenderer struct{}

// FileExtension returns ".json".
func (JSONRenderer) FileExtension() string { return ".json" }

// Render serialises the document to indented JSON.
func (JSONRenderer) Render(doc *Document) (string, error) {
	type fileEntry struct {
		Path    string `json:"path"`
		Content string `json:"content,omitempty"`
		Binary  bool   `json:"binary,omitempty"`
		Skipped string `json:"skipped,omitempty"`
		Bytes   int64  `json:"bytes"`
	}
	type commit struct {
		Hash    string   `json:"hash"`
		Author  string   `json:"author,omitempty"`
		Date    string   `json:"date,omitempty"`
		Message string   `json:"message,omitempty"`
		Files   []string `json:"files,omitempty"`
	}
	type stats struct {
		Files  int   `json:"files"`
		Chars  int   `json:"chars"`
		Tokens int   `json:"tokens"`
		Bytes  int64 `json:"bytes"`
	}
	type payload struct {
		Producer        string      `json:"producer"`
		GeneratedAt     string      `json:"generated_at"`
		HeaderText      string      `json:"header_text,omitempty"`
		DirectoryTree   string      `json:"directory_tree,omitempty"`
		Stats           stats       `json:"stats"`
		Files           []fileEntry `json:"files,omitempty"`
		GitDiffWorkTree string      `json:"git_diff_worktree,omitempty"`
		GitDiffStaged   string      `json:"git_diff_staged,omitempty"`
		GitLog          []commit    `json:"git_log,omitempty"`
		Instruction     string      `json:"instruction,omitempty"`
	}

	p := payload{
		Producer:    doc.AppName + " v" + doc.AppVersion,
		GeneratedAt: doc.GeneratedAt.UTC().Format("2006-01-02T15:04:05Z"),
		HeaderText:  doc.HeaderText,
		Stats: stats{
			Files:  doc.Aggregate.TotalFiles,
			Chars:  doc.Aggregate.TotalChars,
			Tokens: doc.Aggregate.TotalTokens,
			Bytes:  doc.Aggregate.TotalBytes,
		},
		Instruction:     doc.Instruction,
		GitDiffWorkTree: doc.GitDiffWorkTree,
		GitDiffStaged:   doc.GitDiffStaged,
	}
	if doc.Config.Output.DirectoryStructure {
		p.DirectoryTree = doc.DirectoryTree
	}
	if doc.Config.Output.Files {
		for _, f := range doc.Files {
			e := fileEntry{Path: f.RelPath, Bytes: f.Size}
			if f.Skipped {
				e.Skipped = f.SkipReason
				e.Binary = f.IsBinary
			} else {
				e.Content = f.Content
			}
			p.Files = append(p.Files, e)
		}
	}
	for _, c := range doc.GitLog {
		p.GitLog = append(p.GitLog, commit{
			Hash: c.Hash, Author: c.Author, Date: c.Date,
			Message: c.Message, Files: c.Files,
		})
	}
	out, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out) + "\n", nil
}
