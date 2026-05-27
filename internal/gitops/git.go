// Package gitops shells out to the `git` binary to fetch information that is
// relevant to the packed output: working-tree and staged diffs, recent commit
// log, change frequency for sorting, and remote-repo cloning.
//
// Every function in this package is safe to call when git is not installed —
// it returns an error that callers can choose to propagate or silently drop.
package gitops

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ErrNotInstalled is returned when the `git` binary isn't on PATH.
var ErrNotInstalled = errors.New("gitops: git is not installed or not in PATH")

// run executes `git args...` in `dir` and returns trimmed stdout. If git is
// missing it returns ErrNotInstalled.
func run(ctx context.Context, dir string, args ...string) (string, error) {
	if _, err := exec.LookPath("git"); err != nil {
		return "", ErrNotInstalled
	}
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w (%s)", strings.Join(args, " "), err, strings.TrimSpace(errBuf.String()))
	}
	return strings.TrimRight(out.String(), "\n"), nil
}

// IsRepo reports whether `dir` is inside a git work tree.
func IsRepo(dir string) bool {
	if _, err := exec.LookPath("git"); err != nil {
		return false
	}
	out, err := run(nil, dir, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		return false
	}
	return strings.TrimSpace(out) == "true"
}

// DiffWorkTree returns the unstaged diff (working tree against HEAD).
func DiffWorkTree(ctx context.Context, dir string) (string, error) {
	return run(ctx, dir, "diff", "--no-color")
}

// DiffStaged returns the staged diff (index against HEAD).
func DiffStaged(ctx context.Context, dir string) (string, error) {
	return run(ctx, dir, "diff", "--no-color", "--staged")
}

// Commit is a single entry returned by Log.
type Commit struct {
	Hash    string
	Author  string
	Date    string
	Message string
	Files   []string
}

// commitMarker is a sentinel printed at the start of every commit by the
// custom git log format. We use it to slice the output into per-commit
// records without needing the fragile interaction between --name-only and -z.
const commitMarker = "\x01GCCOMMIT\x01"
const logFormat = commitMarker + "%H%x1f%an%x1f%aI%x1f%s"

// Log returns up to `limit` most recent commits with the files each one
// touched. We use a printf-style format with a unique commit-start sentinel
// so commit messages containing newlines don't confuse the parser.
func Log(ctx context.Context, dir string, limit int) ([]Commit, error) {
	if limit <= 0 {
		limit = 50
	}
	out, err := run(ctx, dir,
		"log",
		"--name-only",
		"--max-count", fmt.Sprintf("%d", limit),
		"--pretty=format:"+logFormat,
	)
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	var commits []Commit
	for _, rec := range strings.Split(out, commitMarker) {
		rec = strings.TrimLeft(rec, "\n")
		if rec == "" {
			continue
		}
		// rec is "hash\x1fauthor\x1fdate\x1fsubject\nfile1\nfile2\n..."
		head, files, _ := strings.Cut(rec, "\n")
		parts := strings.Split(head, "\x1f")
		if len(parts) < 4 {
			continue
		}
		commit := Commit{Hash: parts[0], Author: parts[1], Date: parts[2], Message: parts[3]}
		for _, f := range strings.Split(files, "\n") {
			f = strings.TrimSpace(f)
			if f != "" {
				commit.Files = append(commit.Files, f)
			}
		}
		commits = append(commits, commit)
	}
	return commits, nil
}

// ChangeFreq returns per-path change counts based on the last `commits`
// commits. Higher counts indicate files that change more often.
func ChangeFreq(ctx context.Context, dir string, commits int) (map[string]int, error) {
	if commits <= 0 {
		commits = 100
	}
	out, err := run(ctx, dir,
		"log",
		"--name-only",
		"--pretty=format:",
		fmt.Sprintf("--max-count=%d", commits),
	)
	if err != nil {
		return nil, err
	}
	counts := map[string]int{}
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		counts[line]++
	}
	return counts, nil
}

// Clone clones `repoURL` into `dest`. If `branch` is non-empty we try a shallow
// branch clone first and fall back to a full clone + checkout (covering tags
// and arbitrary commit SHAs).
func Clone(ctx context.Context, repoURL, dest, branch string) error {
	if _, err := exec.LookPath("git"); err != nil {
		return ErrNotInstalled
	}
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
	}
	args := []string{"clone", "--depth", "1"}
	if branch != "" {
		args = append(args, "--branch", branch)
	}
	args = append(args, repoURL, dest)
	cmd := exec.CommandContext(ctx, "git", args...)
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err == nil {
		return nil
	}
	// Branch-target may have been a commit SHA. Retry with a full clone + checkout.
	if branch != "" {
		cmd = exec.CommandContext(ctx, "git", "clone", repoURL, dest)
		errBuf.Reset()
		cmd.Stderr = &errBuf
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git clone: %w (%s)", err, strings.TrimSpace(errBuf.String()))
		}
		co := exec.CommandContext(ctx, "git", "checkout", branch)
		co.Dir = dest
		errBuf.Reset()
		co.Stderr = &errBuf
		if err := co.Run(); err != nil {
			return fmt.Errorf("git checkout %s: %w (%s)", branch, err, strings.TrimSpace(errBuf.String()))
		}
		return nil
	}
	return fmt.Errorf("git clone: %s", strings.TrimSpace(errBuf.String()))
}
