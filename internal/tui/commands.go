package tui

import (
	"errors"
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/christophergyman/claude-quick/internal/auth"
	"github.com/christophergyman/claude-quick/internal/devcontainer"
	"github.com/christophergyman/claude-quick/internal/github"
)

// Common errors for nil checks
var (
	errNoInstanceSelected = errors.New("no container instance selected")
	errNoSessionSelected  = errors.New("no tmux session selected")
	errNoWorktreeSelected = errors.New("no worktree selected")
)

// discoverInstances returns a command that discovers devcontainer instances
func (m Model) discoverInstances() tea.Cmd {
	return func() tea.Msg {
		instances := devcontainer.DiscoverInstances(
			m.config.SearchPaths,
			m.config.MaxDepth,
			m.config.ExcludedDirs,
		)
		return instancesDiscoveredMsg{instances: instances}
	}
}

// refreshInstanceStatus returns a command that refreshes container status for all instances
func (m Model) refreshInstanceStatus() tea.Cmd {
	return func() tea.Msg {
		statuses := devcontainer.GetAllInstancesStatus(m.instances)
		return instanceStatusRefreshedMsg{statuses: statuses}
	}
}

// startContainer returns a command that starts the devcontainer
func (m Model) startContainer() tea.Cmd {
	return func() tea.Msg {
		if m.selectedInstance == nil {
			return containerErrorMsg{err: errNoInstanceSelected}
		}

		// Check if devcontainer CLI is available
		if err := devcontainer.CheckCLI(); err != nil {
			return containerErrorMsg{err: err}
		}

		// Resolve and write authentication credentials
		var authWarning string
		if m.config != nil {
			result := m.config.Auth.Resolve(m.selectedInstance.Name)
			if len(result.Credentials) > 0 {
				// Write credentials to file in project directory
				if err := auth.WriteCredentialFile(m.selectedInstance.Path, result.Credentials); err != nil {
					authWarning = fmt.Sprintf("failed to write credentials: %v", err)
				}
			}
			if result.HasErrors() {
				if authWarning != "" {
					authWarning += "; "
				}
				authWarning += result.ErrorSummary()
			}
		}

		// Start the container (path-based, each worktree has unique path)
		if err := devcontainer.Up(m.selectedInstance.Path); err != nil {
			return containerErrorMsg{err: err}
		}

		// Check if tmux is available in container
		if !devcontainer.HasTmux(m.selectedInstance.Path) {
			return containerErrorMsg{err: &tmuxNotFoundError{}}
		}

		return containerStartedMsg{authWarning: authWarning}
	}
}

// stopContainer returns a command that stops the devcontainer
func (m Model) stopContainer() tea.Cmd {
	return func() tea.Msg {
		if m.selectedInstance == nil {
			return containerErrorMsg{err: errNoInstanceSelected}
		}
		if err := devcontainer.Stop(m.selectedInstance.Path); err != nil {
			return containerErrorMsg{err: err}
		}
		// Clean up credential file after stopping container
		auth.CleanupCredentialFile(m.selectedInstance.Path)
		return containerStoppedMsg{}
	}
}

// restartContainer returns a command that restarts the devcontainer
func (m Model) restartContainer() tea.Cmd {
	return func() tea.Msg {
		if m.selectedInstance == nil {
			return containerErrorMsg{err: errNoInstanceSelected}
		}
		if err := devcontainer.Restart(m.selectedInstance.Path); err != nil {
			return containerErrorMsg{err: err}
		}
		return containerRestartedMsg{}
	}
}

// stopTmuxSession returns a command that stops/kills a tmux session
func (m Model) stopTmuxSession() tea.Cmd {
	return func() tea.Msg {
		if m.selectedInstance == nil {
			return containerErrorMsg{err: errNoInstanceSelected}
		}
		if m.selectedSession == nil {
			return containerErrorMsg{err: errNoSessionSelected}
		}
		if err := devcontainer.KillTmuxSession(m.selectedInstance.Path, m.selectedSession.Name); err != nil {
			return containerErrorMsg{err: err}
		}
		return tmuxSessionStoppedMsg{}
	}
}

// restartTmuxSession returns a command that restarts a tmux session (kill + create)
func (m Model) restartTmuxSession() tea.Cmd {
	return func() tea.Msg {
		if m.selectedInstance == nil {
			return containerErrorMsg{err: errNoInstanceSelected}
		}
		if m.selectedSession == nil {
			return containerErrorMsg{err: errNoSessionSelected}
		}
		sessionName := m.selectedSession.Name
		// Kill existing session
		if err := devcontainer.KillTmuxSession(m.selectedInstance.Path, sessionName); err != nil {
			return containerErrorMsg{err: err}
		}
		// Resolve launch command (project-specific or global default)
		launchCmd := m.config.Auth.ResolveLaunchCommand(m.selectedInstance.Name, m.config.LaunchCommand)
		// Create new session with same name
		if err := devcontainer.CreateTmuxSession(m.selectedInstance.Path, sessionName, launchCmd); err != nil {
			return containerErrorMsg{err: err}
		}
		return tmuxSessionRestartedMsg{}
	}
}

// handleContainerStarted is called when container has started
func (m Model) handleContainerStarted() (tea.Model, tea.Cmd) {
	// Transition to loading state with spinner
	m.state = StateLoadingTmuxSessions
	return m, tea.Batch(m.spinner.Tick, m.loadTmuxSessions())
}

// loadTmuxSessions returns a command that loads tmux sessions
func (m Model) loadTmuxSessions() tea.Cmd {
	return func() tea.Msg {
		if m.selectedInstance == nil {
			return containerErrorMsg{err: errNoInstanceSelected}
		}
		sessions, err := devcontainer.ListTmuxSessions(m.selectedInstance.Path)
		if err != nil {
			return containerErrorMsg{err: err}
		}
		return tmuxSessionsLoadedMsg{sessions: sessions}
	}
}

// createTmuxSession creates a new tmux session in the container
func (m Model) createTmuxSession(name string) tea.Cmd {
	return func() tea.Msg {
		if m.selectedInstance == nil {
			return containerErrorMsg{err: errNoInstanceSelected}
		}
		// Resolve launch command (project-specific or global default)
		launchCmd := m.config.Auth.ResolveLaunchCommand(m.selectedInstance.Name, m.config.LaunchCommand)
		if err := devcontainer.CreateTmuxSession(m.selectedInstance.Path, name, launchCmd); err != nil {
			return containerErrorMsg{err: err}
		}
		return tmuxSessionCreatedMsg{}
	}
}

// attachToSession attaches to a tmux session using tea.ExecProcess
// This suspends the TUI, runs tmux as a subprocess, and returns to TUI on detach
func (m Model) attachToSession(sessionName string) (tea.Model, tea.Cmd) {
	if m.selectedInstance == nil {
		m.state = StateError
		m.err = errNoInstanceSelected
		return m, nil
	}
	m.state = StateAttaching

	// Build the command to attach to tmux (path-based)
	c := exec.Command("devcontainer", "exec",
		"--workspace-folder", m.selectedInstance.Path,
		"tmux", "attach", "-t", sessionName)

	// Use tea.ExecProcess to run tmux and return to TUI when done
	return m, tea.ExecProcess(c, func(err error) tea.Msg {
		// This is called when tmux detaches or exits
		return tmuxDetachedMsg{}
	})
}

// deleteWorktree removes the selected git worktree
func (m Model) deleteWorktree() tea.Cmd {
	return func() tea.Msg {
		if m.selectedInstance == nil {
			return containerErrorMsg{err: errNoWorktreeSelected}
		}
		// Pass main repo path to handle cases where worktree directory was deleted externally
		var mainRepoPath string
		if m.selectedInstance.Worktree != nil {
			mainRepoPath = m.selectedInstance.Worktree.MainRepo
		}
		if err := devcontainer.RemoveWorktree(m.selectedInstance.Path, mainRepoPath); err != nil {
			return containerErrorMsg{err: err}
		}
		return worktreeDeletedMsg{}
	}
}

// createWorktree creates a new git worktree with the specified branch
func (m Model) createWorktree(branchName string) tea.Cmd {
	return func() tea.Msg {
		if m.selectedInstance == nil {
			return containerErrorMsg{err: errNoInstanceSelected}
		}
		worktreePath, err := devcontainer.CreateWorktree(m.selectedInstance.Path, branchName)
		if err != nil {
			return containerErrorMsg{err: err}
		}
		return worktreeCreatedMsg{worktreePath: worktreePath}
	}
}

// loadGitHubIssues fetches issues from the current repository
func (m Model) loadGitHubIssues() tea.Cmd {
	return func() tea.Msg {
		if m.selectedInstance == nil {
			return githubIssuesErrorMsg{err: errNoInstanceSelected}
		}

		// Detect repo from git remote
		owner, repo, err := github.DetectRepository(m.selectedInstance.Path)
		if err != nil {
			return githubIssuesErrorMsg{err: err}
		}

		// Fetch issues using gh CLI
		issues, err := github.FetchIssues(owner, repo, m.config.GitHub)
		if err != nil {
			return githubIssuesErrorMsg{err: err}
		}

		return githubIssuesLoadedMsg{
			issues: issues,
			owner:  owner,
			repo:   repo,
		}
	}
}

// loadGitHubIssueDetail fetches the full body of a single issue
func (m Model) loadGitHubIssueDetail() tea.Cmd {
	return func() tea.Msg {
		if m.selectedIssue == nil {
			return githubIssuesErrorMsg{err: errors.New("no issue selected")}
		}

		// Fetch issue body
		body, err := github.FetchIssueBody(m.githubRepoOwner, m.githubRepoName, m.selectedIssue.Number)
		if err != nil {
			return githubIssuesErrorMsg{err: err}
		}

		return githubIssueDetailLoadedMsg{body: body}
	}
}

// createWorktreeFromIssue creates a worktree with auto-generated branch name from issue
func (m Model) createWorktreeFromIssue() tea.Cmd {
	return func() tea.Msg {
		if m.selectedInstance == nil {
			return containerErrorMsg{err: errNoInstanceSelected}
		}
		if m.selectedIssue == nil {
			return containerErrorMsg{err: errors.New("no issue selected")}
		}

		// Generate branch name from issue
		branchName := github.GenerateBranchName(m.selectedIssue, m.config.GitHub.BranchPrefix)

		// Validate branch name (reuse existing validation)
		if err := devcontainer.ValidateBranchName(branchName); err != nil {
			return containerErrorMsg{err: err}
		}

		// Create worktree
		worktreePath, err := devcontainer.CreateWorktree(m.selectedInstance.Path, branchName)
		if err != nil {
			return containerErrorMsg{err: err}
		}

		return githubWorktreeCreatedMsg{
			worktreePath: worktreePath,
			branchName:   branchName,
		}
	}
}
