package tui

import (
	"github.com/christophergyman/claude-quick/internal/devcontainer"
	"github.com/christophergyman/claude-quick/internal/github"
)

// Message types for async operations in Bubbletea.
// These are returned by commands and processed in the Update method.

// instancesDiscoveredMsg is sent when project discovery completes
type instancesDiscoveredMsg struct {
	instances []devcontainer.ContainerInstance
}

// instanceStatusRefreshedMsg is sent when container status refresh completes
type instanceStatusRefreshedMsg struct {
	statuses []devcontainer.ContainerInstanceWithStatus
}

// containerStartedMsg is sent when a container finishes starting
type containerStartedMsg struct {
	// authWarning contains any auth credential resolution warnings (empty if none)
	authWarning string
}

// containerErrorMsg is sent when any container operation fails
type containerErrorMsg struct{ err error }

// tmuxSessionsLoadedMsg is sent when tmux session list is loaded
type tmuxSessionsLoadedMsg struct{ sessions []string }

// tmuxSessionCreatedMsg is sent when a new tmux session is created
type tmuxSessionCreatedMsg struct{}

// containerStoppedMsg is sent when a container is stopped
type containerStoppedMsg struct{}

// containerRestartedMsg is sent when a container is restarted
type containerRestartedMsg struct{}

// tmuxSessionStoppedMsg is sent when a tmux session is killed
type tmuxSessionStoppedMsg struct{}

// tmuxSessionRestartedMsg is sent when a tmux session is restarted
type tmuxSessionRestartedMsg struct{}

// tmuxDetachedMsg is sent when user detaches from tmux
type tmuxDetachedMsg struct{}

// worktreeCreatedMsg is sent when a new git worktree is created
type worktreeCreatedMsg struct {
	worktreePath string
}

// worktreeDeletedMsg is sent when a git worktree is deleted
type worktreeDeletedMsg struct{}

// githubIssuesLoadedMsg is sent when GitHub issues are successfully fetched
type githubIssuesLoadedMsg struct {
	issues []github.Issue
	owner  string
	repo   string
}

// githubIssuesErrorMsg is sent when fetching GitHub issues fails
type githubIssuesErrorMsg struct {
	err error
}

// githubIssueDetailLoadedMsg is sent when a single issue's details are loaded
type githubIssueDetailLoadedMsg struct {
	body string
}

// githubWorktreeCreatedMsg is sent when worktree creation from issue succeeds
type githubWorktreeCreatedMsg struct {
	worktreePath string
	branchName   string
}

// tmuxNotFoundError indicates tmux is not available in the container
type tmuxNotFoundError struct{}

func (e *tmuxNotFoundError) Error() string {
	return "tmux not found in container. Please install tmux in your devcontainer."
}
