package gitops

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// RemoteSpec describes a parsed remote location and any extra metadata
// (subdirectory, branch) embedded in the URL.
type RemoteSpec struct {
	// URL is the canonical clone URL.
	URL string
	// Host (github.com, gitlab.com, etc.).
	Host string
	// Owner/Repo identify the namespace pair where applicable.
	Owner string
	Repo  string
	// Branch is the explicit branch/tag/sha to check out.
	Branch string
	// Subdir is the in-repository directory to focus on, if specified.
	Subdir string
}

// repoOnly matches the convenient `owner/repo` shorthand. The leading
// character is required to be alphanumeric so paths like `./foo` or `../x`
// don't accidentally look like a GitHub shorthand.
var repoOnly = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9._-]*/[A-Za-z0-9_][A-Za-z0-9._-]*$`)

// IsExplicitRemoteURL reports whether `s` looks like a URL we can clone.
// `owner/repo` shorthand returns true so callers can use it as the positional
// argument.
func IsExplicitRemoteURL(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	if repoOnly.MatchString(s) {
		return true
	}
	switch {
	case strings.HasPrefix(s, "http://"),
		strings.HasPrefix(s, "https://"),
		strings.HasPrefix(s, "git://"),
		strings.HasPrefix(s, "ssh://"),
		strings.HasPrefix(s, "git@"):
		return true
	}
	return false
}

// ParseRemote decodes a remote URL, expanding shorthands and extracting any
// `/tree/<ref>/<subdir>` or `/blob/<ref>/<path>` segments embedded in a
// GitHub-style web URL.
func ParseRemote(s string) (RemoteSpec, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return RemoteSpec{}, errors.New("empty remote")
	}
	// Local filesystem path — git can clone from it directly.
	if strings.HasPrefix(s, "/") || strings.HasPrefix(s, "./") || strings.HasPrefix(s, "../") || strings.HasPrefix(s, "file://") {
		return RemoteSpec{URL: s, Host: "local"}, nil
	}
	if repoOnly.MatchString(s) {
		parts := strings.SplitN(s, "/", 2)
		return RemoteSpec{
			URL:   "https://github.com/" + s + ".git",
			Host:  "github.com",
			Owner: parts[0],
			Repo:  parts[1],
		}, nil
	}
	// ssh: git@github.com:owner/repo(.git)?
	if strings.HasPrefix(s, "git@") {
		rest := strings.TrimPrefix(s, "git@")
		host, path, ok := strings.Cut(rest, ":")
		if !ok {
			return RemoteSpec{}, fmt.Errorf("invalid git@ url %q", s)
		}
		path = strings.TrimSuffix(path, ".git")
		owner, repo, _ := strings.Cut(path, "/")
		return RemoteSpec{
			URL:   s,
			Host:  host,
			Owner: owner,
			Repo:  repo,
		}, nil
	}
	u, err := url.Parse(s)
	if err != nil {
		return RemoteSpec{}, err
	}
	if u.Host == "" {
		return RemoteSpec{}, fmt.Errorf("url has no host: %q", s)
	}
	spec := RemoteSpec{Host: u.Host}
	pathSegments := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(pathSegments) >= 2 {
		spec.Owner = pathSegments[0]
		repo := pathSegments[1]
		spec.Repo = strings.TrimSuffix(repo, ".git")
		// GitHub /tree/<ref>/<subdir>... pattern
		if len(pathSegments) >= 4 && (pathSegments[2] == "tree" || pathSegments[2] == "blob") {
			spec.Branch = pathSegments[3]
			if len(pathSegments) >= 5 {
				spec.Subdir = strings.Join(pathSegments[4:], "/")
			}
		}
	}
	// Rebuild a clean clone URL.
	cleanPath := spec.Owner + "/" + spec.Repo + ".git"
	spec.URL = u.Scheme + "://" + u.Host + "/" + cleanPath
	return spec, nil
}
