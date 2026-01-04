package devcontainer

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/christophergyman/claude-quick/internal/constants"
)

// IsGitWorktree checks if the given path is a git worktree and returns its info
// Returns nil if the path is not a git worktree or not a git repository
func IsGitWorktree(path string) *WorktreeInfo {
	gitPath := filepath.Join(path, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		return nil // Not a git repository
	}

	if info.IsDir() {
		// It's a regular git repository (main worktree)
		branch := getGitBranch(path)
		wt := newMainWorktreeInfo(path, branch)
		return &wt
	}

	// .git is a file - this is a worktree
	// Read the gitdir from the file
	content, err := os.ReadFile(gitPath)
	if err != nil {
		return nil
	}

	// Parse: "gitdir: /path/to/.git/worktrees/name"
	line := strings.TrimSpace(string(content))
	if !strings.HasPrefix(line, "gitdir: ") {
		return nil
	}
	gitDir := strings.TrimPrefix(line, "gitdir: ")

	// Find the main repo (go up from .git/worktrees/<name> to .git, then parent)
	// gitDir is like /path/to/main/.git/worktrees/feature-x
	mainGitDir := filepath.Dir(filepath.Dir(gitDir)) // Go up to .git
	mainRepo := filepath.Dir(mainGitDir)             // Go up to main repo

	branch := getGitBranch(path)

	return &WorktreeInfo{
		Path:     path,
		Branch:   branch,
		MainRepo: mainRepo,
		GitDir:   gitDir,
		IsMain:   false,
	}
}

// getGitBranch returns the current branch name for a git repository
func getGitBranch(repoPath string) string {
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return constants.DefaultBranchUnknown
	}
	return strings.TrimSpace(string(output))
}

// ListWorktrees returns all worktrees for a repository (including the main one)
func ListWorktrees(repoPath string) ([]WorktreeInfo, error) {
	// Run git worktree list --porcelain
	cmd := exec.Command("git", "-C", repoPath, "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	var worktrees []WorktreeInfo
	var current WorktreeInfo
	isFirst := true

	// Helper to finalize and append the current worktree
	appendWorktree := func() {
		if current.Path == "" {
			return
		}
		if isFirst {
			worktrees = append(worktrees, newMainWorktreeInfo(current.Path, current.Branch))
		} else {
			worktrees = append(worktrees, newBranchWorktreeInfo(current.Path, current.Branch, repoPath))
		}
		isFirst = false
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			appendWorktree()
			current = WorktreeInfo{}
			continue
		}

		if strings.HasPrefix(line, "worktree ") {
			current.Path = strings.TrimPrefix(line, "worktree ")
		} else if strings.HasPrefix(line, "branch refs/heads/") {
			current.Branch = strings.TrimPrefix(line, "branch refs/heads/")
		} else if strings.HasPrefix(line, "HEAD ") {
			// Detached HEAD - use short SHA as branch name
			if current.Branch == "" {
				sha := strings.TrimPrefix(line, "HEAD ")
				if len(sha) > constants.SHATruncateLength {
					current.Branch = sha[:constants.SHATruncateLength]
				} else {
					current.Branch = sha
				}
			}
		}
	}

	// Handle last worktree if no trailing newline
	appendWorktree()

	return worktrees, nil
}

// GetMainRepo finds the main repository path from any worktree path
func GetMainRepo(worktreePath string) (string, error) {
	info := IsGitWorktree(worktreePath)
	if info == nil {
		return "", fmt.Errorf("not a git repository or worktree")
	}
	return info.MainRepo, nil
}

// CreateWorktree creates a new git worktree with a new branch
// Returns the path to the new worktree directory
func CreateWorktree(repoPath, branchName string) (string, error) {
	// Validate branch name
	if err := ValidateBranchName(branchName); err != nil {
		return "", err
	}

	// Check if this is a git repository
	wtInfo := IsGitWorktree(repoPath)
	if wtInfo == nil {
		return "", fmt.Errorf("not a git repository")
	}

	// Get the main repo path
	mainRepo := wtInfo.MainRepo

	// Create worktree path as sibling directory: repo-branchname
	// Replace "/" with "-" to avoid creating nested directories for hierarchical branches
	repoName := filepath.Base(mainRepo)
	safeBranchName := strings.ReplaceAll(branchName, "/", "-")
	worktreePath := filepath.Join(filepath.Dir(mainRepo), repoName+"-"+safeBranchName)

	// Check if worktree already exists
	if _, err := os.Stat(worktreePath); err == nil {
		return "", fmt.Errorf("worktree directory already exists: %s", worktreePath)
	}

	// Create the worktree with a new branch
	cmd := exec.Command("git", "-C", mainRepo, "worktree", "add", "-b", branchName, worktreePath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to create worktree: %s", stderr.String())
	}

	return worktreePath, nil
}

// RemoveWorktree removes a git worktree
// If mainRepoPath is provided, it will be used when the worktree directory doesn't exist
func RemoveWorktree(worktreePath string, mainRepoPath ...string) error {
	// Stop any running Docker container for this worktree first
	// We attempt this regardless of whether the directory exists
	_ = Stop(worktreePath) // Ignore error - container may not be running

	wtInfo := IsGitWorktree(worktreePath)

	var mainRepo string
	if wtInfo != nil {
		// Directory exists and is a valid worktree
		if wtInfo.IsMain {
			return fmt.Errorf("cannot remove the main worktree")
		}
		mainRepo = wtInfo.MainRepo
	} else {
		// Directory doesn't exist or is not a valid worktree
		// Try to use provided mainRepoPath
		if len(mainRepoPath) > 0 && mainRepoPath[0] != "" {
			mainRepo = mainRepoPath[0]
		} else {
			return fmt.Errorf("not a git worktree")
		}
	}

	// Remove the worktree using --force flag to handle missing directories
	cmd := exec.Command("git", "-C", mainRepo, "worktree", "remove", "--force", worktreePath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// If removal failed, try pruning stale worktrees
		pruneCmd := exec.Command("git", "-C", mainRepo, "worktree", "prune")
		if pruneErr := pruneCmd.Run(); pruneErr == nil {
			// Pruning succeeded, check if the worktree was cleaned up
			return nil
		}
		return fmt.Errorf("failed to remove worktree: %s", stderr.String())
	}

	return nil
}

// ValidateBranchName checks if a branch name is valid for git
func ValidateBranchName(name string) error {
	if name == "" {
		return fmt.Errorf("branch name cannot be empty")
	}
	if constants.IsReservedBranchName(name) {
		return fmt.Errorf("'%s' is a reserved branch name", name)
	}
	// Git branch name rules (simplified)
	if strings.HasPrefix(name, "-") {
		return fmt.Errorf("branch name cannot start with '-'")
	}
	if strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".") {
		return fmt.Errorf("branch name cannot start or end with '.'")
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("branch name cannot contain '..'")
	}
	// Only allow alphanumeric, hyphen, underscore, slash for hierarchical branches
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '/') {
			return fmt.Errorf("branch name contains invalid character: %c", r)
		}
	}
	return nil
}
