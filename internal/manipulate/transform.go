package manipulate

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rajathjn/GopherCram/internal/config"
)

// Apply runs the requested transforms (in a fixed order) on content originally
// loaded from `filename`.
func Apply(filename, content string, opts config.Output) string {
	out := content
	if opts.RemoveComments {
		out = RemoveComments(filename, out)
	}
	if opts.TruncateBase64 {
		out = TruncateBase64(out)
	}
	if opts.Compress {
		out = CompressCode(filename, out)
	}
	if opts.RemoveEmptyLines {
		out = RemoveEmptyLines(out)
	}
	if opts.ShowLineNumbers {
		out = WithLineNumbers(out)
	}
	return out
}

// RemoveEmptyLines drops blank-only lines from content.
func RemoveEmptyLines(s string) string {
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, ln := range lines {
		if strings.TrimSpace(ln) == "" {
			continue
		}
		out = append(out, ln)
	}
	return strings.Join(out, "\n")
}

// WithLineNumbers prefixes each line with its 1-based number, right-aligned in
// a column that fits the largest line number.
func WithLineNumbers(s string) string {
	lines := strings.Split(s, "\n")
	width := digits(len(lines))
	var b strings.Builder
	b.Grow(len(s) + len(lines)*(width+2))
	for i, ln := range lines {
		n := i + 1
		// Avoid printing a synthetic last line if the input ended with `\n`.
		if i == len(lines)-1 && ln == "" {
			break
		}
		// `%*d` style padding without importing fmt is unnecessary here.
		num := numString(n)
		for j := 0; j < width-len(num); j++ {
			b.WriteByte(' ')
		}
		b.WriteString(num)
		b.WriteString(" | ")
		b.WriteString(ln)
		b.WriteByte('\n')
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func digits(n int) int {
	if n <= 0 {
		return 1
	}
	d := 0
	for n > 0 {
		d++
		n /= 10
	}
	return d
}

func numString(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}

// base64Re matches long base64-ish runs (60 chars or more, suffixed with
// optional padding). Catches typical embedded image / certificate blobs.
var base64Re = regexp.MustCompile(`[A-Za-z0-9+/]{60,}={0,2}`)

// TruncateBase64 replaces long base64 payloads with a short placeholder so
// the output stays readable.
func TruncateBase64(s string) string {
	return base64Re.ReplaceAllStringFunc(s, func(m string) string {
		return "[base64 omitted, " + numString(len(m)) + " chars]"
	})
}

// CompressCode performs a coarse "show signatures, hide bodies" transform.
// Lines that look like function/class/method declarations are kept, comments
// and import blocks survive, and long blocks between braces collapse to a
// single placeholder. This is intentionally heuristic — it works well enough
// to dramatically shrink large source files for skimming.
func CompressCode(filename, content string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".go":
		return compressGoLike(content, goSigRe)
	case ".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs":
		return compressGoLike(content, jsSigRe)
	case ".py":
		return compressPython(content)
	case ".java", ".kt", ".cs", ".swift", ".scala", ".cpp", ".cc", ".cxx",
		".c", ".h", ".hpp", ".rs":
		return compressGoLike(content, cLikeSigRe)
	}
	return content
}

var (
	goSigRe     = regexp.MustCompile(`^(?:\s*)(?:package |import |func |type |var |const )`)
	jsSigRe     = regexp.MustCompile(`^(?:\s*)(?:import |export |class |interface |type |function |async function |const |let |var |enum )`)
	cLikeSigRe  = regexp.MustCompile(`^(?:\s*)(?:public |private |protected |internal |static |fn |func |class |struct |trait |interface |enum |impl |namespace |using |import |#include|template)`)
	pyHeaderRe  = regexp.MustCompile(`^(?:\s*)(?:def |class |async def |from |import |@)`)
	indentSpace = regexp.MustCompile(`^(\s*)`)
)

func compressGoLike(content string, sigRe *regexp.Regexp) string {
	lines := strings.Split(content, "\n")
	out := make([]string, 0, len(lines))
	depth := 0
	keepingBody := false
	for _, ln := range lines {
		trim := strings.TrimSpace(ln)
		if depth == 0 {
			// Always keep top-level comments and signatures.
			if trim == "" || strings.HasPrefix(trim, "//") || strings.HasPrefix(trim, "/*") || strings.HasPrefix(trim, "*") {
				out = append(out, ln)
				continue
			}
			if sigRe.MatchString(ln) {
				out = append(out, ln)
				keepingBody = strings.Count(ln, "{") > strings.Count(ln, "}")
			} else {
				out = append(out, ln)
				continue
			}
		}
		depth += strings.Count(ln, "{") - strings.Count(ln, "}")
		if depth < 0 {
			depth = 0
		}
		if depth > 0 && !keepingBody {
			// elide
			continue
		}
		if depth == 0 && keepingBody {
			out = append(out, "  // ...")
			keepingBody = false
		}
	}
	return strings.Join(out, "\n")
}

func compressPython(content string) string {
	lines := strings.Split(content, "\n")
	out := make([]string, 0, len(lines))
	keepIndent := -1
	for _, ln := range lines {
		trim := strings.TrimSpace(ln)
		indent := len(indentSpace.FindString(ln))
		if trim == "" {
			out = append(out, ln)
			continue
		}
		if pyHeaderRe.MatchString(ln) {
			out = append(out, ln)
			if strings.HasSuffix(trim, ":") {
				keepIndent = indent
				out = append(out, strings.Repeat(" ", indent+4)+"# ...")
			}
			continue
		}
		if keepIndent >= 0 && indent > keepIndent {
			continue
		}
		keepIndent = -1
		out = append(out, ln)
	}
	return strings.Join(out, "\n")
}
