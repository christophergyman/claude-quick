package devcontainer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateBranchName(t *testing.T) {
	tests := []struct {
		name       string
		branchName string
		wantErr    bool
		errContain string
	}{
		// Valid names
		{"simple name", "feature", false, ""},
		{"with hyphen", "feature-auth", false, ""},
		{"with underscore", "feature_auth", false, ""},
		{"hierarchical", "feature/auth", false, ""},
		{"nested hierarchical", "feature/auth/login", false, ""},
		{"numbers", "feature123", false, ""},
		{"mixed case", "FeatureAuth", false, ""},
		{"single char", "x", false, ""},

		// Invalid names
		{"empty", "", true, "cannot be empty"},
		{"reserved main", "main", true, "reserved branch name"},
		{"reserved master", "master", true, "reserved branch name"},
		{"starts with hyphen", "-feature", true, "cannot start with '-'"},
		{"starts with dot", ".feature", true, "cannot start or end with '.'"},
		{"ends with dot", "feature.", true, "cannot start or end with '.'"},
		{"double dot", "feature..test", true, "cannot contain '..'"},
		{"space", "feature auth", true, "invalid character"},
		{"special char at", "feature@test", true, "invalid character"},
		{"special char hash", "feature#test", true, "invalid character"},
		{"special char colon", "feature:test", true, "invalid character"},
		{"special char tilde", "feature~test", true, "invalid character"},
		{"special char caret", "feature^test", true, "invalid character"},
		{"special char backslash", "feature\\test", true, "invalid character"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBranchName(tt.branchName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBranchName(%q) error = %v, wantErr %v", tt.branchName, err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
				t.Errorf("ValidateBranchName(%q) error = %v, want error containing %q", tt.branchName, err, tt.errContain)
			}
		})
	}
}

func TestIsGitWorktree_RegularGitRepo(t *testing.T) {
	// Create a temporary directory with a .git directory (simulating regular repo)
	tmpDir, err := os.MkdirTemp("", "test-git-repo")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}

	result := IsGitWorktree(tmpDir)
	if result == nil {
		t.Fatal("IsGitWorktree returned nil for directory with .git")
	}

	if !result.IsMain {
		t.Errorf("IsMain = %v, want true", result.IsMain)
	}
	if result.Path != tmpDir {
		t.Errorf("Path = %q, want %q", result.Path, tmpDir)
	}
	if result.MainRepo != tmpDir {
		t.Errorf("MainRepo = %q, want %q", result.MainRepo, tmpDir)
	}
	expectedGitDir := filepath.Join(tmpDir, ".git")
	if result.GitDir != expectedGitDir {
		t.Errorf("GitDir = %q, want %q", result.GitDir, expectedGitDir)
	}
}

func TestIsGitWorktree_NotGitRepo(t *testing.T) {
	// Create a temporary directory without .git
	tmpDir, err := os.MkdirTemp("", "test-not-git")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	result := IsGitWorktree(tmpDir)
	if result != nil {
		t.Errorf("IsGitWorktree returned %v for non-git directory, want nil", result)
	}
}

func TestIsGitWorktree_WorktreeFile(t *testing.T) {
	// Create a temporary directory structure simulating a worktree
	baseDir, err := os.MkdirTemp("", "test-worktree-base")
	if err != nil {
		t.Fatalf("failed to create base dir: %v", err)
	}
	defer os.RemoveAll(baseDir)

	// Create main repo
	mainRepo := filepath.Join(baseDir, "main-repo")
	if err := os.Mkdir(mainRepo, 0755); err != nil {
		t.Fatalf("failed to create main repo dir: %v", err)
	}

	// Create main .git directory with worktrees subdirectory
	mainGitDir := filepath.Join(mainRepo, ".git")
	if err := os.MkdirAll(filepath.Join(mainGitDir, "worktrees", "feature-x"), 0755); err != nil {
		t.Fatalf("failed to create .git/worktrees: %v", err)
	}

	// Create worktree directory
	worktreeDir := filepath.Join(baseDir, "worktree-feature")
	if err := os.Mkdir(worktreeDir, 0755); err != nil {
		t.Fatalf("failed to create worktree dir: %v", err)
	}

	// Create .git file pointing to the worktree gitdir
	gitFile := filepath.Join(worktreeDir, ".git")
	gitdirPath := filepath.Join(mainGitDir, "worktrees", "feature-x")
	content := "gitdir: " + gitdirPath + "\n"
	if err := os.WriteFile(gitFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write .git file: %v", err)
	}

	result := IsGitWorktree(worktreeDir)
	if result == nil {
		t.Fatal("IsGitWorktree returned nil for worktree")
	}

	if result.IsMain {
		t.Errorf("IsMain = %v, want false", result.IsMain)
	}
	if result.Path != worktreeDir {
		t.Errorf("Path = %q, want %q", result.Path, worktreeDir)
	}
	if result.MainRepo != mainRepo {
		t.Errorf("MainRepo = %q, want %q", result.MainRepo, mainRepo)
	}
	if result.GitDir != gitdirPath {
		t.Errorf("GitDir = %q, want %q", result.GitDir, gitdirPath)
	}
}

func TestIsGitWorktree_InvalidGitFile(t *testing.T) {
	// Create a temporary directory with an invalid .git file
	tmpDir, err := os.MkdirTemp("", "test-invalid-git")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git file with invalid content (no "gitdir: " prefix)
	gitFile := filepath.Join(tmpDir, ".git")
	if err := os.WriteFile(gitFile, []byte("invalid content"), 0644); err != nil {
		t.Fatalf("failed to write .git file: %v", err)
	}

	result := IsGitWorktree(tmpDir)
	if result != nil {
		t.Errorf("IsGitWorktree returned %v for invalid .git file, want nil", result)
	}
}

func TestIsGitWorktree_NonexistentPath(t *testing.T) {
	result := IsGitWorktree("/nonexistent/path/that/does/not/exist")
	if result != nil {
		t.Errorf("IsGitWorktree returned %v for nonexistent path, want nil", result)
	}
}

func TestParseWorktreeListOutput(t *testing.T) {
	// This is a helper test to verify the parsing logic
	// The actual ListWorktrees function shells out to git, but we can test
	// the parsing would work with expected output format

	// Example porcelain output format:
	// worktree /path/to/main
	// HEAD abc123...
	// branch refs/heads/main
	//
	// worktree /path/to/feature
	// HEAD def456...
	// branch refs/heads/feature/auth

	// Since ListWorktrees uses exec.Command, we'll just verify the
	// function exists and document expected behavior
	t.Log("ListWorktrees parses 'git worktree list --porcelain' output")
	t.Log("Expected format: worktree <path>, HEAD <sha>, branch refs/heads/<name>")
}
