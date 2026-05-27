package gitops

import "testing"

func TestIsExplicitRemoteURL(t *testing.T) {
	cases := map[string]bool{
		"":                              false,
		"https://github.com/o/r":        true,
		"http://x.example/repo.git":     true,
		"git@github.com:o/r.git":        true,
		"git://x/repo":                  true,
		"ssh://x/repo":                  true,
		"owner/repo":                    true,
		"./local":                       false,
		"not a url":                     false,
		"https://gitlab.com/g/sub/repo": true,
	}
	for in, want := range cases {
		if got := IsExplicitRemoteURL(in); got != want {
			t.Errorf("IsExplicitRemoteURL(%q)=%v, want %v", in, got, want)
		}
	}
}

func TestParseRemote_Shorthand(t *testing.T) {
	spec, err := ParseRemote("owner/repo")
	if err != nil {
		t.Fatal(err)
	}
	if spec.Owner != "owner" || spec.Repo != "repo" {
		t.Errorf("unexpected: %+v", spec)
	}
	if spec.URL != "https://github.com/owner/repo.git" {
		t.Errorf("URL=%s", spec.URL)
	}
}

func TestParseRemote_SSH(t *testing.T) {
	spec, err := ParseRemote("git@github.com:owner/repo.git")
	if err != nil {
		t.Fatal(err)
	}
	if spec.Host != "github.com" || spec.Owner != "owner" || spec.Repo != "repo" {
		t.Errorf("unexpected: %+v", spec)
	}
}

func TestParseRemote_HTTPSWithBranch(t *testing.T) {
	spec, err := ParseRemote("https://github.com/owner/repo/tree/main/sub/dir")
	if err != nil {
		t.Fatal(err)
	}
	if spec.Branch != "main" || spec.Subdir != "sub/dir" {
		t.Errorf("unexpected: %+v", spec)
	}
}

func TestParseRemote_Empty(t *testing.T) {
	if _, err := ParseRemote(""); err == nil {
		t.Error("expected error")
	}
}

func TestParseRemote_BadSSH(t *testing.T) {
	if _, err := ParseRemote("git@nohost"); err == nil {
		t.Error("expected error for malformed git@")
	}
}

func TestIsRepo_NotARepo(t *testing.T) {
	dir := t.TempDir()
	if IsRepo(dir) {
		t.Error("temp dir is not a git repo")
	}
}
