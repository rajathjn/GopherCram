package gitops

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func needGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
}

func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	mustRun(t, dir, "git", "init", "-q")
	mustRun(t, dir, "git", "config", "user.email", "t@example.test")
	mustRun(t, dir, "git", "config", "user.name", "Test")
	mustRun(t, dir, "git", "config", "commit.gpgsign", "false")
	return dir
}

func mustRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%v: %v\n%s", args, err, out)
	}
}

func TestRunMissing(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		// emulate "git missing" — skip would also work
		t.Skip("git missing — cannot verify the success path")
	}
	if _, err := run(context.Background(), t.TempDir(), "nope-cmd-arg"); err == nil {
		t.Error("expected error for bogus git subcommand")
	}
}

func TestIsRepoAndLog(t *testing.T) {
	needGit(t)
	dir := initRepo(t)
	if IsRepo(dir) != true {
		t.Fatal("expected initial repo to be detected")
	}
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustRun(t, dir, "git", "add", "a.txt")
	mustRun(t, dir, "git", "commit", "-q", "-m", "initial")

	commits, err := Log(context.Background(), dir, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 1 || commits[0].Message != "initial" || len(commits[0].Files) != 1 {
		t.Errorf("unexpected commits: %+v", commits)
	}

	cf, err := ChangeFreq(context.Background(), dir, 5)
	if err != nil {
		t.Fatal(err)
	}
	if cf["a.txt"] != 1 {
		t.Errorf("changefreq=%+v", cf)
	}
}

func TestDiffs(t *testing.T) {
	needGit(t)
	dir := initRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustRun(t, dir, "git", "add", "a.txt")
	mustRun(t, dir, "git", "commit", "-q", "-m", "init")
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hi changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	diff, err := DiffWorkTree(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if diff == "" {
		t.Error("expected non-empty diff")
	}
	mustRun(t, dir, "git", "add", "a.txt")
	staged, err := DiffStaged(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if staged == "" {
		t.Error("expected staged diff")
	}
}
