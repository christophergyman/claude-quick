package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/christophergyman/claude-quick/internal/devcontainer"
)

// handleKeyPress processes keyboard input based on current state
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case StateDashboard:
		return m.handleDashboardKey(msg)
	case StateConfirmStop, StateConfirmRestart:
		return m.handleConfirmKey(msg)
	case StateConfirmDeleteWorktree:
		return m.handleConfirmDeleteWorktreeKey(msg)
	case StateConfirmTmuxStop, StateConfirmTmuxRestart:
		return m.handleTmuxConfirmKey(msg)
	case StateTmuxSelect:
		return m.handleTmuxSelectKey(msg)
	case StateNewSessionInput:
		return m.handleNewSessionInputKey(msg)
	case StateNewWorktreeInput:
		return m.handleNewWorktreeInputKey(msg)
	case StateGitHubIssuesList:
		return m.handleGitHubIssuesListKey(msg)
	case StateGitHubIssueDetail:
		return m.handleGitHubIssueDetailKey(msg)
	case StateError:
		// Any key returns to container select
		m.state = StateDashboard
		m.err = nil
		return m, nil
	case StateShowConfig:
		// Any key returns to previous state
		m.state = m.previousState
		return m, nil
	}
	return m, nil
}

func (m Model) handleDashboardKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.instancesStatus)-1 {
			m.cursor++
		}

	case "enter":
		if len(m.instancesStatus) > 0 {
			m.selectedInstance = &m.instancesStatus[m.cursor].ContainerInstance
			if m.instancesStatus[m.cursor].Status == devcontainer.StatusRunning {
				// Container is running, load tmux sessions
				m.state = StateLoadingTmuxSessions
				return m, tea.Batch(m.spinner.Tick, m.loadTmuxSessions())
			}
			// Container is stopped or unknown, start it
			m.state = StateContainerStarting
			return m, tea.Batch(
				m.spinner.Tick,
				m.startContainer(),
			)
		}

	case "x":
		if len(m.instancesStatus) > 0 {
			m.selectedInstance = &m.instancesStatus[m.cursor].ContainerInstance
			m.state = StateConfirmStop
		}

	case "r":
		if len(m.instancesStatus) > 0 {
			m.selectedInstance = &m.instancesStatus[m.cursor].ContainerInstance
			m.state = StateConfirmRestart
		}

	case "R":
		// Manual refresh
		m.state = StateRefreshingStatus
		return m, tea.Batch(m.spinner.Tick, m.refreshInstanceStatus())

	case "n":
		// Create new worktree - requires selecting a git project first
		if len(m.instancesStatus) > 0 {
			selected := &m.instancesStatus[m.cursor].ContainerInstance
			// Only allow creating worktrees for git repositories
			if selected.Worktree == nil {
				m.state = StateError
				m.err = fmt.Errorf("cannot create worktree: not a git repository")
				m.errHint = "Press any key to go back"
				return m, nil
			}
			m.selectedInstance = selected
			m.state = StateNewWorktreeInput
			m.worktreeInput.SetValue("")
			m.worktreeInput.Focus()
			return m, textinput.Blink
		}

	case "d":
		// Delete worktree - only for non-main worktrees
		if len(m.instancesStatus) > 0 {
			selected := &m.instancesStatus[m.cursor].ContainerInstance
			// Only allow deleting non-main worktrees
			if selected.Worktree == nil {
				m.state = StateError
				m.err = fmt.Errorf("cannot delete: not a git worktree")
				m.errHint = "Press any key to go back"
				return m, nil
			}
			if selected.Worktree.IsMain {
				m.state = StateError
				m.err = fmt.Errorf("cannot delete the main worktree")
				m.errHint = "Press any key to go back"
				return m, nil
			}
			m.selectedInstance = selected
			m.state = StateConfirmDeleteWorktree
		}

	case "?":
		m.previousState = m.state
		m.state = StateShowConfig
		return m, nil

	case "t":
		// Toggle dark/light theme
		m.darkMode = !m.darkMode
		ApplyTheme(m.darkMode)
		return m, nil

	case "g":
		// Open GitHub Issues - requires selecting a git project first
		if len(m.instancesStatus) > 0 {
			selected := &m.instancesStatus[m.cursor].ContainerInstance
			// Only allow for git repositories
			if selected.Worktree == nil {
				m.state = StateError
				m.err = fmt.Errorf("cannot open GitHub issues: not a git repository")
				m.errHint = "Press any key to go back"
				return m, nil
			}
			m.selectedInstance = selected
			m.state = StateGitHubIssuesLoading
			return m, tea.Batch(m.spinner.Tick, m.loadGitHubIssues())
		}
	}
	return m, nil
}

func (m Model) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if m.state == StateConfirmStop {
			m.state = StateContainerStopping
			return m, tea.Batch(m.spinner.Tick, m.stopContainer())
		}
		m.state = StateContainerRestarting
		return m, tea.Batch(m.spinner.Tick, m.restartContainer())
	case "n", "N", "esc":
		m.state = StateDashboard
		m.selectedInstance = nil
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleTmuxConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if m.state == StateConfirmTmuxStop {
			m.state = StateTmuxStopping
			return m, tea.Batch(m.spinner.Tick, m.stopTmuxSession())
		}
		m.state = StateTmuxRestarting
		return m, tea.Batch(m.spinner.Tick, m.restartTmuxSession())
	case "n", "N", "esc":
		m.state = StateTmuxSelect
		m.selectedSession = nil
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleTmuxSelectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	totalOptions := TotalTmuxOptions(m.tmuxSessions)

	switch msg.String() {
	case "q", "esc":
		// Go back to container select with status refresh
		m.state = StateRefreshingStatus
		m.cursor = 0
		m.selectedInstance = nil
		return m, tea.Batch(m.spinner.Tick, m.refreshInstanceStatus())

	case "ctrl+c":
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < totalOptions-1 {
			m.cursor++
		}

	case "x":
		// Stop/kill selected tmux session (only for existing sessions)
		if m.cursor < len(m.tmuxSessions) {
			m.selectedSession = &m.tmuxSessions[m.cursor]
			m.state = StateConfirmTmuxStop
		}

	case "r":
		// Restart selected tmux session (only for existing sessions)
		if m.cursor < len(m.tmuxSessions) {
			m.selectedSession = &m.tmuxSessions[m.cursor]
			m.state = StateConfirmTmuxRestart
		}

	case "?":
		m.previousState = m.state
		m.state = StateShowConfig
		return m, nil

	case "t":
		// Toggle dark/light theme
		m.darkMode = !m.darkMode
		ApplyTheme(m.darkMode)
		return m, nil

	case "enter":
		if IsNewSessionSelected(m.tmuxSessions, m.cursor) {
			// Show text input for new session name
			m.state = StateNewSessionInput
			m.textInput.SetValue("")
			m.textInput.Focus()
			return m, textinput.Blink
		}
		// Attach to existing session
		if m.cursor < len(m.tmuxSessions) {
			return m.attachToSession(m.tmuxSessions[m.cursor].Name)
		}
	}
	return m, nil
}

func (m Model) handleNewSessionInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel and go back to tmux select
		m.state = StateTmuxSelect
		return m, nil

	case "ctrl+c":
		return m, tea.Quit

	case "enter":
		name := m.textInput.Value()
		if name == "" {
			name = m.config.DefaultSessionName
		}
		m.state = StateAttaching
		return m, tea.Batch(
			m.spinner.Tick,
			m.createTmuxSession(name),
		)
	}

	// Pass other keys to text input
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m Model) handleNewWorktreeInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel and go back to dashboard
		m.state = StateDashboard
		return m, nil

	case "ctrl+c":
		return m, tea.Quit

	case "enter":
		branchName := m.worktreeInput.Value()
		if branchName == "" {
			m.state = StateError
			m.err = devcontainer.ValidateBranchName("")
			m.errHint = "Press any key to go back"
			return m, nil
		}
		if err := devcontainer.ValidateBranchName(branchName); err != nil {
			m.state = StateError
			m.err = err
			m.errHint = "Press any key to go back"
			return m, nil
		}
		m.state = StateCreatingWorktree
		return m, tea.Batch(
			m.spinner.Tick,
			m.createWorktree(branchName),
		)
	}

	// Pass other keys to text input
	var cmd tea.Cmd
	m.worktreeInput, cmd = m.worktreeInput.Update(msg)
	return m, cmd
}

func (m Model) handleConfirmDeleteWorktreeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		m.state = StateDeletingWorktree
		return m, tea.Batch(m.spinner.Tick, m.deleteWorktree())
	case "n", "N", "esc":
		m.state = StateDashboard
		m.selectedInstance = nil
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleGitHubIssuesListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		// Go back to dashboard
		m.state = StateDashboard
		m.githubIssues = nil
		m.selectedIssue = nil
		m.cursor = 0
		return m, nil

	case "ctrl+c":
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.githubIssues)-1 {
			m.cursor++
		}

	case "r":
		// Refresh issues
		m.state = StateGitHubIssuesLoading
		return m, tea.Batch(m.spinner.Tick, m.loadGitHubIssues())

	case "enter":
		// Create worktree from selected issue
		if len(m.githubIssues) > 0 && m.cursor < len(m.githubIssues) {
			m.selectedIssue = &m.githubIssues[m.cursor]
			m.state = StateGitHubWorktreeCreating
			return m, tea.Batch(
				m.spinner.Tick,
				m.createWorktreeFromIssue(),
			)
		}

	case "v":
		// View issue details
		if len(m.githubIssues) > 0 && m.cursor < len(m.githubIssues) {
			m.selectedIssue = &m.githubIssues[m.cursor]
			m.state = StateGitHubIssueDetailLoading
			return m, tea.Batch(m.spinner.Tick, m.loadGitHubIssueDetail())
		}

	case "t":
		// Toggle theme
		m.darkMode = !m.darkMode
		ApplyTheme(m.darkMode)
		return m, nil
	}
	return m, nil
}

func (m Model) handleGitHubIssueDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		// Go back to issues list
		m.state = StateGitHubIssuesList
		return m, nil

	case "ctrl+c":
		return m, tea.Quit

	case "enter":
		// Create worktree from this issue
		if m.selectedIssue != nil {
			m.state = StateGitHubWorktreeCreating
			return m, tea.Batch(
				m.spinner.Tick,
				m.createWorktreeFromIssue(),
			)
		}

	case "t":
		// Toggle theme
		m.darkMode = !m.darkMode
		ApplyTheme(m.darkMode)
		return m, nil
	}
	return m, nil
}
