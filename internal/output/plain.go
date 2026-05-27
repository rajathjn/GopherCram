package output

import "strings"

// PlainRenderer emits a plain-text document with simple ASCII separators.
type PlainRenderer struct{}

// FileExtension returns ".txt".
func (PlainRenderer) FileExtension() string { return ".txt" }

const plainSep = "================================================================"

// Render produces the plain-text representation of the document.
func (PlainRenderer) Render(doc *Document) (string, error) {
	var b strings.Builder
	if doc.Config.Output.FileSummary {
		b.WriteString(plainSep + "\n")
		b.WriteString("SUMMARY\n")
		b.WriteString(plainSep + "\n")
		b.WriteString("Producer:     " + doc.AppName + " v" + doc.AppVersion + "\n")
		b.WriteString("Generated at: " + doc.GeneratedAt.UTC().Format("2006-01-02 15:04:05 UTC") + "\n")
		b.WriteString("Files:        " + itoa(doc.Aggregate.TotalFiles) + "\n")
		b.WriteString("Characters:   " + itoa(doc.Aggregate.TotalChars) + "\n")
		b.WriteString("Tokens (~):   " + itoa(doc.Aggregate.TotalTokens) + "\n")
		b.WriteString("Bytes:        " + itoa64(doc.Aggregate.TotalBytes) + "\n\n")
	}
	if doc.HeaderText != "" {
		b.WriteString(plainSep + "\n")
		b.WriteString("USER HEADER\n")
		b.WriteString(plainSep + "\n")
		b.WriteString(doc.HeaderText)
		b.WriteString("\n\n")
	}
	if doc.Config.Output.DirectoryStructure {
		b.WriteString(plainSep + "\n")
		b.WriteString("DIRECTORY STRUCTURE\n")
		b.WriteString(plainSep + "\n")
		b.WriteString(doc.DirectoryTree)
		b.WriteString("\n\n")
	}
	if doc.Config.Output.Files && len(doc.Files) > 0 {
		b.WriteString(plainSep + "\n")
		b.WriteString("FILES\n")
		b.WriteString(plainSep + "\n\n")
		for _, f := range doc.Files {
			if f.Skipped {
				continue
			}
			b.WriteString("---- " + f.RelPath + " ")
			pad := 70 - len(f.RelPath)
			if pad > 0 {
				b.WriteString(strings.Repeat("-", pad))
			}
			b.WriteString("\n")
			b.WriteString(f.Content)
			if !strings.HasSuffix(f.Content, "\n") {
				b.WriteString("\n")
			}
			b.WriteString("\n")
		}
	}
	if doc.GitDiffWorkTree != "" {
		b.WriteString(plainSep + "\n")
		b.WriteString("GIT DIFF (WORKING TREE)\n")
		b.WriteString(plainSep + "\n")
		b.WriteString(doc.GitDiffWorkTree)
		b.WriteString("\n\n")
	}
	if doc.GitDiffStaged != "" {
		b.WriteString(plainSep + "\n")
		b.WriteString("GIT DIFF (STAGED)\n")
		b.WriteString(plainSep + "\n")
		b.WriteString(doc.GitDiffStaged)
		b.WriteString("\n\n")
	}
	if len(doc.GitLog) > 0 {
		b.WriteString(plainSep + "\n")
		b.WriteString("GIT LOG\n")
		b.WriteString(plainSep + "\n")
		for _, c := range doc.GitLog {
			b.WriteString(c.Hash + " " + c.Date + " — " + strings.SplitN(c.Message, "\n", 2)[0] + "\n")
		}
		b.WriteString("\n")
	}
	if doc.Instruction != "" {
		b.WriteString(plainSep + "\n")
		b.WriteString("INSTRUCTION\n")
		b.WriteString(plainSep + "\n")
		b.WriteString(doc.Instruction)
		b.WriteString("\n")
	}
	return b.String(), nil
}
