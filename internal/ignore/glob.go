// Package ignore provides gitignore-style pattern matching with `**` support.
//
// The matcher in this package is intentionally small: it implements the subset
// of glob semantics needed for source-tree filtering, namely:
//
//   - `*`  matches any run of non-separator characters
//   - `**` matches any run of characters including separators
//   - `?`  matches a single non-separator character
//   - `[abc]` and `[a-z]` character classes
//   - trailing `/` requires the path to be a directory
//   - leading `!` negates a previously-matching rule
//   - patterns without `/` match against any segment of the path
//
// A path is considered ignored when the final non-negated rule matches.
package ignore

import (
	"strings"
)

// Match reports whether `path` (slash-separated, relative) matches `pattern`.
// `isDir` is used to honour directory-only patterns (those ending in "/").
func Match(pattern, path string, isDir bool) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" || strings.HasPrefix(pattern, "#") {
		return false
	}

	if strings.HasPrefix(pattern, "!") {
		// negation handled at PatternSet level; treat the rest of the pattern
		// as the matching body here.
		pattern = pattern[1:]
	}

	dirOnly := false
	if strings.HasSuffix(pattern, "/") {
		dirOnly = true
		pattern = strings.TrimSuffix(pattern, "/")
	}
	if dirOnly && !isDir {
		return false
	}

	// Normalize: a pattern containing no "/" except possibly leading should
	// be matched against every segment of the path.
	anchored := strings.HasPrefix(pattern, "/")
	if anchored {
		pattern = strings.TrimPrefix(pattern, "/")
	}
	containsSlash := strings.ContainsRune(pattern, '/')

	if anchored || containsSlash {
		return matchPath(pattern, path)
	}

	// Floating pattern — match against the full path and every suffix that
	// starts at a directory boundary.
	if matchPath(pattern, path) {
		return true
	}
	idx := 0
	for {
		i := strings.IndexByte(path[idx:], '/')
		if i < 0 {
			break
		}
		idx += i + 1
		if matchPath(pattern, path[idx:]) {
			return true
		}
	}
	return false
}

// matchPath performs a recursive-descent match between a pattern that may
// include "**" and a slash-separated path.
func matchPath(pattern, name string) bool {
	pi, ni := 0, 0
	// star/double-star backtracking marks
	starPi, starNi := -1, -1
	doublePi, doubleNi := -1, -1
	for ni < len(name) {
		if pi < len(pattern) {
			c := pattern[pi]
			switch c {
			case '*':
				if pi+1 < len(pattern) && pattern[pi+1] == '*' {
					// `**` consumes any characters (including '/').
					// Skip optional trailing '/'.
					pi += 2
					if pi < len(pattern) && pattern[pi] == '/' {
						pi++
					}
					doublePi = pi
					doubleNi = ni
					continue
				}
				// single `*` matches non-separator chars
				starPi = pi + 1
				starNi = ni
				pi++
				continue
			case '?':
				if name[ni] != '/' {
					pi++
					ni++
					continue
				}
			case '[':
				ok, advance := matchClass(pattern[pi:], name[ni])
				if ok && name[ni] != '/' {
					pi += advance
					ni++
					continue
				}
			case '\\':
				if pi+1 < len(pattern) {
					pi++
					if pattern[pi] == name[ni] {
						pi++
						ni++
						continue
					}
				}
			default:
				if c == name[ni] {
					pi++
					ni++
					continue
				}
			}
		}
		// failed match — try backtracking
		if starPi != -1 && name[starNi] != '/' {
			pi = starPi
			starNi++
			ni = starNi
			continue
		}
		if doublePi != -1 {
			doubleNi++
			pi = doublePi
			ni = doubleNi
			continue
		}
		return false
	}
	// remaining pattern must be only `*` or `**`
	for pi < len(pattern) {
		if pattern[pi] == '*' {
			pi++
			continue
		}
		return false
	}
	return true
}

// matchClass matches a `[...]` character class against c and returns whether
// it matched along with the number of pattern bytes consumed.
func matchClass(pat string, c byte) (bool, int) {
	// pat[0] == '['
	i := 1
	neg := false
	if i < len(pat) && (pat[i] == '!' || pat[i] == '^') {
		neg = true
		i++
	}
	matched := false
	for i < len(pat) && pat[i] != ']' {
		lo := pat[i]
		i++
		if i+1 < len(pat) && pat[i] == '-' && pat[i+1] != ']' {
			hi := pat[i+1]
			i += 2
			if c >= lo && c <= hi {
				matched = true
			}
		} else if c == lo {
			matched = true
		}
	}
	if i < len(pat) && pat[i] == ']' {
		i++
	}
	if neg {
		matched = !matched
	}
	return matched, i
}
