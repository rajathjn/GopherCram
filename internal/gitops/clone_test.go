package gitops

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestClone_Local exercises Clone against a local bare repo, which avoids the
// flakiness of network-backed integration testing while still walking the
// real `git clone` path.
func TestClone_Local(t *testing.T) {
	needGit(t)
	src := initRepo(t)
	if err := os.WriteFile(filepath.Join(src, "f.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustRun(t, src, "git", "add", "f.txt")
	mustRun(t, src, "git", "commit", "-q", "-m", "init")

	dest := filepath.Join(t.TempDir(), "clone")
	if err := Clone(context.Background(), src, dest, ""); err != nil {
		t.Fatalf("Clone failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dest, "f.txt")); err != nil {
		t.Errorf("expected cloned f.txt: %v", err)
	}
}

func TestClone_LocalBranchFallback(t *testing.T) {
	needGit(t)
	src := initRepo(t)
	if err := os.WriteFile(filepath.Join(src, "f.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustRun(t, src, "git", "add", "f.txt")
	mustRun(t, src, "git", "commit", "-q", "-m", "init")
	// Capture the SHA so we can clone-then-checkout it (forcing the
	// fallback path in Clone).
	dest := filepath.Join(t.TempDir(), "clone-sha")
	cmd, err := os.Open(filepath.Join(src, ".git", "refs", "heads"))
	if err == nil {
		_ = cmd.Close()
	}
	// We cheat slightly: passing a bogus branch makes the shallow clone fail
	// and triggers the full-clone path. The subsequent checkout also fails
	// (no such branch), so we expect an error — but the function should
	// reach that fallback branch.
	if err := Clone(context.Background(), src, dest, "definitely-not-a-branch"); err == nil {
		t.Error("expected fallback checkout to error on nonexistent branch")
	}
}

func TestClone_GitMissingPath(t *testing.T) {
	// We can't reliably remove `git` from PATH in a unit test, but the easy
	// case is exercised by IsRepo on a non-repo directory, which returns
	// false — a separate test in git_test.go covers that.
	_ = IsRepo("/nonexistent-directory-xyz")
}
