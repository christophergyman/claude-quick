package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/christophergyman/claude-quick/internal/config"
	"github.com/christophergyman/claude-quick/internal/constants"
	"github.com/christophergyman/claude-quick/internal/devcontainer"
	"github.com/christophergyman/claude-quick/internal/github"
	"github.com/christophergyman/claude-quick/internal/tmux"
)

// Model is the main Bubbletea model for the TUI application
type Model struct {
	state            State
	instances        []devcontainer.ContainerInstance
	instancesStatus  []devcontainer.ContainerInstanceWithStatus
	selectedInstance *devcontainer.ContainerInstance
	tmuxSessions     []tmux.Session
	selectedSession  *tmux.Session
	cursor           int
	spinner          spinner.Model
	textInput        textinput.Model
	worktreeInput    textinput.Model
	err              error
	errHint          string
	width            int
	height           int
	config           *config.Config
	previousState    State
	authWarning      string // Warning message if auth credentials failed to resolve
	darkMode         bool   // Current theme mode (true = dark, false = light)

	// GitHub Issues state
	githubIssues    []github.Issue  // Cached list of issues
	selectedIssue   *github.Issue   // Currently selected issue
	githubRepoOwner string          // Detected owner (e.g., "christophergyman")
	githubRepoName  string          // Detected repo name (e.g., "claude-quick")

	// Auto-start state (for GitHub issue worktree creation)
	pendingAutoStart      bool   // Whether to auto-start after discovery
	autoStartWorktreePath string // Path of newly created worktree to auto-start
}

// getInstanceName safely returns the selected instance display name
func (m Model) getInstanceName() string {
	if m.selectedInstance == nil {
		return ""
	}
	return m.selectedInstance.DisplayName()
}

// getSessionName safely returns the selected session name
func (m Model) getSessionName() string {
	if m.selectedSession == nil {
		return ""
	}
	return m.selectedSession.Name
}

// getWorktreeBranch safely returns the selected worktree's branch name
func (m Model) getWorktreeBranch() string {
	if m.selectedInstance == nil || m.selectedInstance.Worktree == nil {
		return ""
	}
	return m.selectedInstance.Worktree.Branch
}

// newTextInput creates a configured text input with the given placeholder
func newTextInput(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = constants.TextInputCharLimit
	ti.Width = constants.TextInputWidth
	return ti
}

// New creates a new Model with discovered instances
func New(instances []devcontainer.ContainerInstance, cfg *config.Config) Model {
	// Initialize theme from config
	darkMode := cfg.IsDarkMode()
	ApplyTheme(darkMode)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	return Model{
		state:         StateDashboard,
		instances:     instances,
		spinner:       s,
		textInput:     newTextInput(cfg.DefaultSessionName),
		worktreeInput: newTextInput(constants.DefaultWorktreePlaceholder),
		config:        cfg,
		darkMode:      darkMode,
	}
}

// NewWithDiscovery creates a Model that will discover instances asynchronously
func NewWithDiscovery(cfg *config.Config) Model {
	// Initialize theme from config
	darkMode := cfg.IsDarkMode()
	ApplyTheme(darkMode)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	return Model{
		state:         StateDiscovering,
		instances:     nil,
		spinner:       s,
		textInput:     newTextInput(cfg.DefaultSessionName),
		worktreeInput: newTextInput(constants.DefaultWorktreePlaceholder),
		config:        cfg,
		darkMode:      darkMode,
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	if m.state == StateDiscovering {
		return tea.Batch(m.spinner.Tick, m.discoverInstances())
	}
	return nil
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case instancesDiscoveredMsg:
		m.instances = msg.instances
		m.state = StateRefreshingStatus
		m.cursor = 0
		return m, tea.Batch(m.spinner.Tick, m.refreshInstanceStatus())

	case instanceStatusRefreshedMsg:
		m.instancesStatus = msg.statuses

		// Check if we need to auto-start a newly created worktree
		if m.pendingAutoStart && m.autoStartWorktreePath != "" {
			m.pendingAutoStart = false
			// Find the instance matching the worktree path
			for i, status := range m.instancesStatus {
				if status.Path == m.autoStartWorktreePath {
					m.selectedInstance = &m.instancesStatus[i].ContainerInstance
					m.cursor = i
					m.autoStartWorktreePath = ""
					// Start the container
					m.state = StateContainerStarting
					return m, tea.Batch(m.spinner.Tick, m.startContainer())
				}
			}
			// If not found, clear and go to dashboard
			m.autoStartWorktreePath = ""
		}

		m.state = StateDashboard
		return m, nil

	case tmuxDetachedMsg:
		// User detached from tmux, return to dashboard with status refresh
		m.state = StateRefreshingStatus
		m.selectedInstance = nil
		m.cursor = 0
		return m, tea.Batch(m.spinner.Tick, m.refreshInstanceStatus())

	case containerStartedMsg:
		m.authWarning = msg.authWarning
		return m.handleContainerStarted()

	case containerErrorMsg:
		m.state = StateError
		m.err = msg.err
		m.errHint = "Press any key to go back"
		return m, nil

	case tmuxSessionsLoadedMsg:
		m.tmuxSessions = tmux.ParseSessions(msg.sessions)
		m.state = StateTmuxSelect
		m.cursor = 0
		return m, nil

	case tmuxSessionCreatedMsg:
		// Session created, now attach
		sessionName := m.textInput.Value()
		return m.attachToSession(sessionName)

	case containerStoppedMsg, containerRestartedMsg:
		// Refresh status after container operation
		m.state = StateRefreshingStatus
		m.selectedInstance = nil
		return m, tea.Batch(m.spinner.Tick, m.refreshInstanceStatus())

	case tmuxSessionStoppedMsg, tmuxSessionRestartedMsg:
		// Reload tmux sessions after stop/restart with loading animation
		m.selectedSession = nil
		m.cursor = 0
		m.state = StateLoadingTmuxSessions
		return m, tea.Batch(m.spinner.Tick, m.loadTmuxSessions())

	case worktreeCreatedMsg:
		// Worktree created, refresh instances
		m.state = StateDiscovering
		return m, tea.Batch(m.spinner.Tick, m.discoverInstances())

	case worktreeDeletedMsg:
		// Worktree deleted, refresh instances
		m.state = StateDiscovering
		m.selectedInstance = nil
		return m, tea.Batch(m.spinner.Tick, m.discoverInstances())

	case githubIssuesLoadedMsg:
		m.githubIssues = msg.issues
		m.githubRepoOwner = msg.owner
		m.githubRepoName = msg.repo
		m.state = StateGitHubIssuesList
		m.cursor = 0
		return m, nil

	case githubIssuesErrorMsg:
		m.state = StateError
		m.err = msg.err
		m.errHint = "Press any key to go back"
		return m, nil

	case githubIssueDetailLoadedMsg:
		// Update selected issue with body
		if m.selectedIssue != nil {
			m.selectedIssue.Body = msg.body
		}
		m.state = StateGitHubIssueDetail
		return m, nil

	case githubWorktreeCreatedMsg:
		// Worktree created from issue, refresh and auto-start
		m.githubIssues = nil
		m.selectedIssue = nil
		// Set up auto-start for after discovery completes
		m.pendingAutoStart = true
		m.autoStartWorktreePath = msg.worktreePath
		m.state = StateDiscovering
		return m, tea.Batch(m.spinner.Tick, m.discoverInstances())
	}

	// Update text input if in input state
	if m.state == StateNewSessionInput {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	// Update worktree input if in worktree name input state
	if m.state == StateNewWorktreeInput {
		var cmd tea.Cmd
		m.worktreeInput, cmd = m.worktreeInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View implements tea.Model
func (m Model) View() string {
	switch m.state {
	case StateDiscovering:
		return RenderDiscovering(m.spinner.View())

	case StateRefreshingStatus:
		return RenderRefreshingStatus(m.spinner.View())

	case StateDashboard:
		return RenderDashboard(m.instancesStatus, m.cursor, m.width)

	case StateContainerStarting:
		return RenderContainerStarting(m.getInstanceName(), m.spinner.View())

	case StateConfirmStop:
		return RenderConfirmDialog("stop", m.getInstanceName())

	case StateConfirmRestart:
		return RenderConfirmDialog("restart", m.getInstanceName())

	case StateContainerStopping:
		return RenderContainerOperation("Stopping", m.getInstanceName(), m.spinner.View())

	case StateContainerRestarting:
		return RenderContainerOperation("Restarting", m.getInstanceName(), m.spinner.View())

	case StateConfirmTmuxStop:
		return RenderTmuxConfirmDialog("stop", m.getSessionName())

	case StateConfirmTmuxRestart:
		return RenderTmuxConfirmDialog("restart", m.getSessionName())

	case StateTmuxStopping:
		return RenderTmuxOperation("Stopping", m.getSessionName(), m.spinner.View())

	case StateTmuxRestarting:
		return RenderTmuxOperation("Restarting", m.getSessionName(), m.spinner.View())

	case StateLoadingTmuxSessions:
		return RenderLoadingTmuxSessions(m.getInstanceName(), m.spinner.View())

	case StateTmuxSelect:
		return RenderTmuxSelect(m.getInstanceName(), m.tmuxSessions, m.cursor, m.authWarning)

	case StateNewSessionInput:
		return RenderNewSessionInput(m.getInstanceName(), m.textInput)

	case StateAttaching:
		sessionName := m.textInput.Value()
		if m.cursor < len(m.tmuxSessions) {
			sessionName = m.tmuxSessions[m.cursor].Name
		}
		return RenderAttaching(m.getInstanceName(), sessionName, m.spinner.View())

	case StateNewWorktreeInput:
		projectName := ""
		if m.selectedInstance != nil {
			projectName = m.selectedInstance.Name
		}
		return RenderNewWorktreeInput(projectName, m.worktreeInput)

	case StateCreatingWorktree:
		return RenderCreatingWorktree(m.worktreeInput.Value(), m.spinner.View())

	case StateConfirmDeleteWorktree:
		return RenderConfirmDeleteWorktree(m.getWorktreeBranch())

	case StateDeletingWorktree:
		return RenderDeletingWorktree(m.getWorktreeBranch(), m.spinner.View())

	case StateError:
		return RenderError(m.err, m.errHint)

	case StateShowConfig:
		return RenderConfigDisplay(m.config)

	case StateGitHubIssuesLoading:
		return RenderGitHubIssuesLoading(m.spinner.View())

	case StateGitHubIssuesList:
		return RenderGitHubIssuesList(m.githubIssues, m.cursor, m.githubRepoOwner, m.githubRepoName, m.width)

	case StateGitHubIssueDetailLoading:
		issueNum := 0
		if m.selectedIssue != nil {
			issueNum = m.selectedIssue.Number
		}
		return RenderGitHubIssueDetailLoading(issueNum, m.spinner.View())

	case StateGitHubIssueDetail:
		body := ""
		if m.selectedIssue != nil {
			body = m.selectedIssue.Body
		}
		return RenderGitHubIssueDetail(m.selectedIssue, body, m.width)

	case StateGitHubWorktreeCreating:
		issueNum := 0
		if m.selectedIssue != nil {
			issueNum = m.selectedIssue.Number
		}
		return RenderGitHubWorktreeCreating(issueNum, m.spinner.View())
	}

	return ""
}
