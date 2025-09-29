package gittool

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetBranchesIncludesCreatedBranch(t *testing.T) {
	repoDir, cleanup := initGitRepo(t)
	t.Cleanup(cleanup)

	runGitCmd(t, repoDir, "checkout", "-b", "feature")

	withWorkdir(t, repoDir, func() {
		branches := GetBranches()
		if !contains(branches, "main") {
			t.Fatalf("GetBranches() missing main branch: %v", branches)
		}
		if !contains(branches, "feature") {
			t.Fatalf("GetBranches() missing feature branch: %v", branches)
		}
	})
}

func TestDiffReturnsChangesFromForkPoint(t *testing.T) {
	repoDir, cleanup := initGitRepo(t)
	t.Cleanup(cleanup)

	runGitCmd(t, repoDir, "checkout", "-b", "feature")

	filePath := filepath.Join(repoDir, "readme.txt")
	if err := os.WriteFile(filePath, []byte("beta\n"), 0o644); err != nil {
		t.Fatalf("write feature content: %v", err)
	}
	runGitCmd(t, repoDir, "add", "readme.txt")
	runGitCmd(t, repoDir, "commit", "-m", "replace content")

	withWorkdir(t, repoDir, func() {
		diff := Diff("main")
		if diff == "" {
			t.Fatal("Diff() returned empty string")
		}
		if !strings.Contains(diff, "+beta") || !strings.Contains(diff, "-alpha") {
			t.Fatalf("Diff() missing expected hunks: %s", diff)
		}
	})
}

func initGitRepo(t *testing.T) (string, func()) {
	t.Helper()

	repoDir := t.TempDir()

	runGitCmd(t, repoDir, "init")

	filePath := filepath.Join(repoDir, "readme.txt")
	if err := os.WriteFile(filePath, []byte("alpha\n"), 0o644); err != nil {
		t.Fatalf("write initial content: %v", err)
	}

	runGitCmd(t, repoDir, "add", "readme.txt")
	runGitCmd(t, repoDir, "commit", "-m", "initial commit")
	runGitCmd(t, repoDir, "branch", "-M", "main")

	return repoDir, func() {}
}

func runGitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Test User",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test User",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)

	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func contains(list []string, target string) bool {
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}

func withWorkdir(t *testing.T, dir string, fn func()) {
	t.Helper()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir to %s: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})

	fn()
}

func TestDiffNoChangesReturnsEmpty(t *testing.T) {
	repoDir, cleanup := initGitRepo(t)
	t.Cleanup(cleanup)

	withWorkdir(t, repoDir, func() {
		diff := Diff("main")
		if diff != "" {
			t.Fatalf("Diff() expected empty string, got %q", diff)
		}
	})
}

func TestRunGitVersion(t *testing.T) {
	out, err := runGit("--version")
	if err != nil {
		t.Fatalf("runGit(--version) error: %v", err)
	}

	if !strings.HasPrefix(out, "git version") {
		t.Fatalf("runGit(--version) = %q, want prefix git version", out)
	}
}

func TestRunGitErrorIncludesStderr(t *testing.T) {
	_, err := runGit("definitely-not-a-git-command")
	if err == nil {
		t.Fatalf("runGit() expected error for invalid command")
	}

	msg := err.Error()
	if !strings.Contains(msg, "definitely-not-a-git-command") {
		t.Fatalf("runGit() error %q missing command name", msg)
	}
}
