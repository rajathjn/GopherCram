package ignore

import "testing"

func TestMatchBasic(t *testing.T) {
	cases := []struct {
		pattern, path string
		want          bool
	}{
		{"*.go", "main.go", true},
		{"*.go", "src/main.go", true}, // floating
		{"/*.go", "main.go", true},
		{"/*.go", "src/main.go", false},
		{"src/*.go", "src/main.go", true},
		{"src/*.go", "src/sub/main.go", false},
		{"src/**/*.go", "src/sub/main.go", true},
		{"src/**/*.go", "src/main.go", true},
		{"node_modules", "x/node_modules", true},
		{"**/node_modules/**", "a/node_modules/b/c.js", true},
		{"foo?bar", "foozbar", true},
		{"foo?bar", "foo/bar", false},
		{"[abc].txt", "a.txt", true},
		{"[abc].txt", "d.txt", false},
		{"[!abc].txt", "d.txt", true},
		{"[a-z].txt", "m.txt", true},
		{"[A-Z].txt", "m.txt", false},
		{"", "anything", false},
		{"# comment", "x", false},
	}
	for _, c := range cases {
		got := Match(c.pattern, c.path, false)
		if got != c.want {
			t.Errorf("Match(%q,%q)=%v, want %v", c.pattern, c.path, got, c.want)
		}
	}
}

func TestMatchDirOnly(t *testing.T) {
	if Match("dist/", "dist/file.js", true) == false {
		// the rule applies to directories; here "dist/file.js" is a file path
		// but the rule matches the dir portion at the start. Since `Match`
		// receives a single path with isDir for that path, we expect false
		// when isDir=false. Verify both forms.
	}
	if Match("dist/", "dist", true) != true {
		t.Error("dist/ should match dist dir")
	}
	if Match("dist/", "dist", false) != false {
		t.Error("dist/ should not match dist file")
	}
}

func TestMatchEscape(t *testing.T) {
	if !Match(`\*.txt`, "*.txt", false) {
		t.Error("escape should match literal *")
	}
}

func TestMatchTrailingStar(t *testing.T) {
	if !Match("foo*", "foo", false) {
		t.Error("foo* should match foo")
	}
	if !Match("foo**", "foo/bar/baz", false) {
		t.Error("foo** should match foo/bar/baz")
	}
}

func TestMatchClassBoundaries(t *testing.T) {
	ok, n := matchClass("[abc]rest", 'b')
	if !ok || n != 5 {
		t.Errorf("matchClass abc/b: ok=%v n=%d", ok, n)
	}
	ok, _ = matchClass("[^a-c]rest", 'z')
	if !ok {
		t.Error("negated class should match z")
	}
	ok, _ = matchClass("[a-c]", 'b')
	if !ok {
		t.Error("range should match b")
	}
}
