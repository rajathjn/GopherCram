package ignore

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// rule is a parsed pattern entry from one source (e.g. a single line of a
// .gitignore file).
type rule struct {
	pattern  string
	negate   bool
	dirOnly  bool
	anchored bool
}

// PatternSet is an ordered collection of rules. The last-matching rule wins,
// which mirrors gitignore semantics.
type PatternSet struct {
	rules []rule
}

// New constructs an empty PatternSet.
func New() *PatternSet { return &PatternSet{} }

// Add appends a single pattern line, parsing the leading `!` and trailing `/`
// markers. Blank lines and comments are ignored.
func (s *PatternSet) Add(line string) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return
	}
	r := rule{}
	if strings.HasPrefix(line, "!") {
		r.negate = true
		line = line[1:]
	}
	if strings.HasSuffix(line, "/") {
		r.dirOnly = true
		line = strings.TrimSuffix(line, "/")
	}
	if strings.HasPrefix(line, "/") {
		r.anchored = true
		line = strings.TrimPrefix(line, "/")
	}
	if line == "" {
		return
	}
	r.pattern = line
	s.rules = append(s.rules, r)
}

// AddAll appends multiple patterns at once.
func (s *PatternSet) AddAll(lines []string) {
	for _, line := range lines {
		s.Add(line)
	}
}

// LoadFile parses a gitignore-style file into the set. Missing files are
// silently ignored.
func (s *PatternSet) LoadFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer f.Close()
	return s.LoadReader(f)
}

// LoadReader parses gitignore-style rules from a Reader.
func (s *PatternSet) LoadReader(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		s.Add(scanner.Text())
	}
	return scanner.Err()
}

// Len returns the number of rules in the set.
func (s *PatternSet) Len() int { return len(s.rules) }

// MatchesPath returns whether `path` (slash-separated, relative to the
// pattern set's anchor) should be ignored.
func (s *PatternSet) MatchesPath(path string, isDir bool) bool {
	if s == nil || len(s.rules) == 0 {
		return false
	}
	// gitignore semantics: last matching pattern wins; a negation can re-include
	// a previously-ignored path.
	ignored := false
	path = filepath.ToSlash(path)
	for _, r := range s.rules {
		if r.dirOnly && !isDir {
			continue
		}
		var hit bool
		if r.anchored || strings.ContainsRune(r.pattern, '/') {
			hit = matchPath(r.pattern, path)
		} else {
			hit = matchAnySegment(r.pattern, path)
		}
		if hit {
			ignored = !r.negate
		}
	}
	return ignored
}

func matchAnySegment(pattern, path string) bool {
	if matchPath(pattern, path) {
		return true
	}
	idx := 0
	for {
		i := strings.IndexByte(path[idx:], '/')
		if i < 0 {
			return false
		}
		idx += i + 1
		if matchPath(pattern, path[idx:]) {
			return true
		}
	}
}

// Merge appends the rules of `other` onto this set. The original ordering is
// preserved so that gitignore's "last match wins" rule remains predictable.
func (s *PatternSet) Merge(other *PatternSet) {
	if other == nil {
		return
	}
	s.rules = append(s.rules, other.rules...)
}
