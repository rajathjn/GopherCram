package output

import (
	"path/filepath"
	"strings"
)

// MarkdownRenderer produces a single Markdown document.
type MarkdownRenderer struct{}

// FileExtension returns ".md".
func (MarkdownRenderer) FileExtension() string { return ".md" }

// Render returns the document as Markdown.
func (MarkdownRenderer) Render(doc *Document) (string, error) {
	var b strings.Builder
	if doc.Config.Output.FileSummary {
		writeMarkdownSummary(&b, doc)
	}
	if doc.HeaderText != "" {
		b.WriteString("## User Header\n\n")
		b.WriteString(doc.HeaderText)
		b.WriteString("\n\n")
	}
	if doc.Config.Output.DirectoryStructure {
		b.WriteString("## Directory Structure\n\n")
		b.WriteString("```\n")
		b.WriteString(doc.DirectoryTree)
		b.WriteString("\n```\n\n")
	}
	if doc.Config.Output.Files && len(doc.Files) > 0 {
		b.WriteString("## Files\n\n")
		for _, f := range doc.Files {
			if f.Skipped {
				continue
			}
			b.WriteString("### `")
			b.WriteString(f.RelPath)
			b.WriteString("`\n\n")
			fence := pickFence(f.Content)
			b.WriteString(fence)
			b.WriteString(languageFromPath(f.RelPath))
			b.WriteString("\n")
			b.WriteString(f.Content)
			if !strings.HasSuffix(f.Content, "\n") {
				b.WriteString("\n")
			}
			b.WriteString(fence)
			b.WriteString("\n\n")
		}
	}
	if doc.GitDiffWorkTree != "" || doc.GitDiffStaged != "" {
		b.WriteString("## Git Diffs\n\n")
		if doc.GitDiffWorkTree != "" {
			b.WriteString("### Working Tree\n\n```diff\n")
			b.WriteString(doc.GitDiffWorkTree)
			b.WriteString("\n```\n\n")
		}
		if doc.GitDiffStaged != "" {
			b.WriteString("### Staged\n\n```diff\n")
			b.WriteString(doc.GitDiffStaged)
			b.WriteString("\n```\n\n")
		}
	}
	if len(doc.GitLog) > 0 {
		b.WriteString("## Git Log\n\n")
		for _, c := range doc.GitLog {
			b.WriteString("- `")
			b.WriteString(c.Hash)
			b.WriteString("` ")
			b.WriteString(c.Date)
			b.WriteString(" — ")
			b.WriteString(strings.SplitN(c.Message, "\n", 2)[0])
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	if doc.Instruction != "" {
		b.WriteString("## Instruction\n\n")
		b.WriteString(doc.Instruction)
		b.WriteString("\n")
	}
	return b.String(), nil
}

func writeMarkdownSummary(b *strings.Builder, doc *Document) {
	b.WriteString("# ")
	b.WriteString(doc.AppName)
	b.WriteString(" Output\n\n")
	b.WriteString("> Packed by **")
	b.WriteString(doc.AppName)
	b.WriteString("** v")
	b.WriteString(doc.AppVersion)
	b.WriteString(" at ")
	b.WriteString(doc.GeneratedAt.UTC().Format("2006-01-02 15:04:05 UTC"))
	b.WriteString("\n\n")
	b.WriteString("| Metric | Value |\n|---|---|\n")
	b.WriteString("| Files | " + itoa(doc.Aggregate.TotalFiles) + " |\n")
	b.WriteString("| Characters | " + itoa(doc.Aggregate.TotalChars) + " |\n")
	b.WriteString("| Tokens (approx) | " + itoa(doc.Aggregate.TotalTokens) + " |\n")
	b.WriteString("| Bytes | " + itoa64(doc.Aggregate.TotalBytes) + " |\n\n")
}

// pickFence returns a fence (```` ``` ```` or longer) that does not collide
// with any fence inside content.
func pickFence(content string) string {
	max := 3
	for _, line := range strings.Split(content, "\n") {
		i := 0
		for i < len(line) && line[i] == '`' {
			i++
		}
		if i >= max {
			max = i + 1
		}
	}
	return strings.Repeat("`", max)
}

// languageFromPath maps common file extensions to Markdown info-string
// language hints. Unknown extensions return "" (raw block).
func languageFromPath(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".js", ".cjs", ".mjs":
		return "javascript"
	case ".jsx":
		return "jsx"
	case ".ts":
		return "typescript"
	case ".tsx":
		return "tsx"
	case ".rb":
		return "ruby"
	case ".rs":
		return "rust"
	case ".java":
		return "java"
	case ".kt", ".kts":
		return "kotlin"
	case ".swift":
		return "swift"
	case ".c", ".h":
		return "c"
	case ".cpp", ".cc", ".cxx", ".hpp":
		return "cpp"
	case ".cs":
		return "csharp"
	case ".php":
		return "php"
	case ".sh", ".bash":
		return "bash"
	case ".zsh":
		return "zsh"
	case ".fish":
		return "fish"
	case ".yml", ".yaml":
		return "yaml"
	case ".toml":
		return "toml"
	case ".json":
		return "json"
	case ".xml":
		return "xml"
	case ".html", ".htm":
		return "html"
	case ".css":
		return "css"
	case ".scss":
		return "scss"
	case ".sql":
		return "sql"
	case ".md", ".markdown":
		return "markdown"
	case ".lua":
		return "lua"
	case ".dart":
		return "dart"
	case ".scala":
		return "scala"
	case ".pl", ".pm":
		return "perl"
	case ".r":
		return "r"
	}
	return ""
}
