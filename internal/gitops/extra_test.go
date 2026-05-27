package gitops

import (
	"context"
	"path/filepath"
	"testing"
)

func TestChangeFreq_NotARepo(t *testing.T) {
	needGit(t)
	dir := t.TempDir()
	if _, err := ChangeFreq(context.Background(), dir, 5); err == nil {
		t.Error("expected error from non-repo")
	}
}

func TestChangeFreq_DefaultLimit(t *testing.T) {
	needGit(t)
	dir := initRepo(t)
	// Create one commit so the command succeeds.
	mustRun(t, dir, "git", "commit", "--allow-empty", "-q", "-m", "x")
	cf, err := ChangeFreq(context.Background(), dir, 0)
	if err != nil {
		t.Fatal(err)
	}
	// Empty commit, no files changed — counts may be empty.
	_ = cf
}

func TestLog_DefaultLimit(t *testing.T) {
	needGit(t)
	dir := initRepo(t)
	mustRun(t, dir, "git", "commit", "--allow-empty", "-q", "-m", "x")
	if _, err := Log(context.Background(), dir, 0); err != nil {
		t.Fatal(err)
	}
}

func TestLog_NoCommits(t *testing.T) {
	needGit(t)
	dir := initRepo(t)
	// On an empty repo `git log` exits 128 — our wrapper surfaces that as an
	// error. Either an error or an empty slice is acceptable behaviour.
	commits, err := Log(context.Background(), dir, 5)
	if err == nil && len(commits) != 0 {
		t.Errorf("expected zero commits on empty repo, got %d", len(commits))
	}
}

func TestClone_NonexistentSource(t *testing.T) {
	needGit(t)
	dest := filepath.Join(t.TempDir(), "dest")
	if err := Clone(context.Background(), "/definitely/does/not/exist", dest, ""); err == nil {
		t.Error("expected clone failure")
	}
}

func TestDiff_NotARepo(t *testing.T) {
	needGit(t)
	dir := t.TempDir()
	if _, err := DiffWorkTree(context.Background(), dir); err == nil {
		t.Error("expected diff error in non-repo")
	}
	if _, err := DiffStaged(context.Background(), dir); err == nil {
		t.Error("expected staged diff error in non-repo")
	}
}
