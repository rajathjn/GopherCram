// Package manipulate transforms file content prior to output. It supports
// stripping comments, removing blank lines, truncating long base64 payloads,
// and performing a lightweight code-compression pass that keeps signatures
// and discards bodies.
package manipulate

import (
	"path/filepath"
	"regexp"
	"strings"
)

// RemoveComments strips comments from text using a language-aware lexer when
// possible, falling back to a regex-based heuristic for unrecognised file
// types. The original blank lines are preserved.
func RemoveComments(filename, content string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".go", ".c", ".cpp", ".cc", ".cxx", ".h", ".hpp", ".java", ".js", ".jsx",
		".ts", ".tsx", ".kt", ".kts", ".swift", ".scala", ".rs", ".cs", ".dart",
		".php", ".m", ".mm":
		return stripCStyleComments(content)
	case ".py", ".pyw", ".rb", ".sh", ".bash", ".zsh", ".fish", ".pl", ".pm",
		".r", ".yaml", ".yml", ".toml", ".dockerfile", ".makefile":
		return stripHashComments(content)
	case ".lua":
		return stripLuaComments(content)
	case ".sql":
		return stripSQLComments(content)
	case ".html", ".htm", ".xml", ".svg", ".vue":
		return stripXMLComments(content)
	case ".css", ".scss", ".less":
		return stripCStyleComments(content)
	default:
		// Best-effort fallback: try C-style first; it covers most cases.
		return stripCStyleComments(content)
	}
}

// stripCStyleComments removes // line and /* */ block comments, respecting
// string and character literals so we don't mangle code containing "//" inside
// strings.
func stripCStyleComments(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	const (
		code = iota
		strDouble
		strSingle
		strBacktick
		lineComment
		blockComment
	)
	state := code
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch state {
		case code:
			switch c {
			case '/':
				if i+1 < len(s) && s[i+1] == '/' {
					state = lineComment
					i++
					continue
				}
				if i+1 < len(s) && s[i+1] == '*' {
					state = blockComment
					i++
					continue
				}
				b.WriteByte(c)
			case '"':
				state = strDouble
				b.WriteByte(c)
			case '\'':
				state = strSingle
				b.WriteByte(c)
			case '`':
				state = strBacktick
				b.WriteByte(c)
			default:
				b.WriteByte(c)
			}
		case strDouble:
			b.WriteByte(c)
			if c == '\\' && i+1 < len(s) {
				b.WriteByte(s[i+1])
				i++
				continue
			}
			if c == '"' {
				state = code
			}
		case strSingle:
			b.WriteByte(c)
			if c == '\\' && i+1 < len(s) {
				b.WriteByte(s[i+1])
				i++
				continue
			}
			if c == '\'' {
				state = code
			}
		case strBacktick:
			b.WriteByte(c)
			if c == '`' {
				state = code
			}
		case lineComment:
			if c == '\n' {
				b.WriteByte(c)
				state = code
			}
		case blockComment:
			if c == '*' && i+1 < len(s) && s[i+1] == '/' {
				state = code
				i++
			} else if c == '\n' {
				// preserve newlines so line numbers stay aligned
				b.WriteByte(c)
			}
		}
	}
	return b.String()
}

func stripHashComments(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	const (
		code = iota
		strDouble
		strSingle
		comment
	)
	state := code
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch state {
		case code:
			switch c {
			case '#':
				state = comment
			case '"':
				state = strDouble
				b.WriteByte(c)
			case '\'':
				state = strSingle
				b.WriteByte(c)
			default:
				b.WriteByte(c)
			}
		case strDouble:
			b.WriteByte(c)
			if c == '\\' && i+1 < len(s) {
				b.WriteByte(s[i+1])
				i++
				continue
			}
			if c == '"' {
				state = code
			}
		case strSingle:
			b.WriteByte(c)
			if c == '\\' && i+1 < len(s) {
				b.WriteByte(s[i+1])
				i++
				continue
			}
			if c == '\'' {
				state = code
			}
		case comment:
			if c == '\n' {
				b.WriteByte(c)
				state = code
			}
		}
	}
	return b.String()
}

func stripLuaComments(s string) string {
	// `--` line and `--[[ ]]` block; we use a simple pass that doesn't try to
	// handle nested long brackets.
	re := regexp.MustCompile(`(?s)--\[\[.*?]]|--[^\n]*`)
	return re.ReplaceAllStringFunc(s, func(m string) string {
		// preserve newlines inside multi-line comments
		return strings.Repeat("\n", strings.Count(m, "\n"))
	})
}

func stripSQLComments(s string) string {
	// `--` line and `/* */` block.
	out := regexp.MustCompile(`--[^\n]*`).ReplaceAllString(s, "")
	out = regexp.MustCompile(`(?s)/\*.*?\*/`).ReplaceAllStringFunc(out, func(m string) string {
		return strings.Repeat("\n", strings.Count(m, "\n"))
	})
	return out
}

func stripXMLComments(s string) string {
	re := regexp.MustCompile(`(?s)<!--.*?-->`)
	return re.ReplaceAllStringFunc(s, func(m string) string {
		return strings.Repeat("\n", strings.Count(m, "\n"))
	})
}
