// Package devcontainer provides functionality for managing devcontainers,
// git worktrees, and tmux sessions within devcontainers.
package devcontainer

import "path/filepath"

// Project represents a devcontainer project
type Project struct {
	Name string // Project directory name
	Path string // Full path to the project (workspace folder)
}

// WorktreeInfo represents a git worktree
type WorktreeInfo struct {
	Path     string // Worktree directory path
	Branch   string // Current branch name
	MainRepo string // Path to main repository
	GitDir   string // Path to worktree gitdir (.git/worktrees/<name>)
	IsMain   bool   // True if this is the main worktree
}

// newMainWorktreeInfo creates a WorktreeInfo for the main worktree of a repo
func newMainWorktreeInfo(path, branch string) WorktreeInfo {
	return WorktreeInfo{
		Path:     path,
		Branch:   branch,
		MainRepo: path,
		GitDir:   filepath.Join(path, ".git"),
		IsMain:   true,
	}
}

// newBranchWorktreeInfo creates a WorktreeInfo for a non-main worktree
func newBranchWorktreeInfo(path, branch, mainRepo string) WorktreeInfo {
	worktreeName := filepath.Base(path)
	return WorktreeInfo{
		Path:     path,
		Branch:   branch,
		MainRepo: mainRepo,
		GitDir:   filepath.Join(mainRepo, ".git", "worktrees", worktreeName),
		IsMain:   false,
	}
}

// ContainerStatus represents the runtime status of a devcontainer
type ContainerStatus string

const (
	StatusRunning ContainerStatus = "running"
	StatusStopped ContainerStatus = "stopped"
	StatusUnknown ContainerStatus = "unknown"
)

// ContainerInstance represents a specific devcontainer instance
// Each instance corresponds to a main repo or a git worktree
type ContainerInstance struct {
	Project              // Embedded: Name and Path (workspace folder)
	ConfigPath string    // Full path to devcontainer.json (from main repo)
	Worktree   *WorktreeInfo // Worktree info (nil for main repo if not a worktree)
}

// ContainerInstanceWithStatus extends ContainerInstance with runtime info
type ContainerInstanceWithStatus struct {
	ContainerInstance
	Status       ContainerStatus
	ContainerID  string
	SessionCount int
}

// DisplayName returns the formatted name for UI display
func (c ContainerInstance) DisplayName() string {
	if c.Worktree != nil && !c.Worktree.IsMain {
		return c.Name + " [" + c.Worktree.Branch + "]"
	}
	return c.Name
}
