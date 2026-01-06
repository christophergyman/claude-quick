package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/christophergyman/claude-quick/internal/auth"
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
	warning          string // Warning message (auth, push failures, etc.)
	darkMode         bool   // Current theme mode (true = dark, false = light)

	// GitHub Issues state
	githubIssues    []github.Issue  // Cached list of issues
	selectedIssue   *github.Issue   // Currently selected issue
	githubRepoOwner string          // Detected owner (e.g., "christophergyman")
	githubRepoName  string          // Detected repo name (e.g., "claude-quick")

	// Auto-start state (for GitHub issue worktree creation)
	pendingAutoStart      bool   // Whether to auto-start after discovery
	autoStartWorktreePath string // Path of newly created worktree to auto-start

	// Wizard state for guided configuration setup
	wizardSearchPaths   []string          // Editable search paths list
	wizardPathInput     textinput.Model   // Input for adding search paths
	wizardCredSource    auth.SourceType   // Selected credential source (file/env/command)
	wizardCredValue     textinput.Model   // Credential value input
	wizardCredentials   []auth.Credential // Credentials being configured
	wizardSessionInput  textinput.Model   // Default session name input
	wizardTimeoutInput  textinput.Model   // Container timeout input
	wizardLaunchInput   textinput.Model   // Launch command input
	wizardMaxDepthInput textinput.Model   // Max search depth input
	wizardDarkMode      bool              // Dark mode toggle
	wizardCursor        int               // Cursor for list navigation in wizard
	wizardEditMode      bool              // Whether currently editing a field
	wizardPathWarnings  map[string]bool   // Map of path -> exists (false means warning)
	wizardFromDashboard bool              // Whether wizard was launched from dashboard
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

// NewWithWizard creates a Model that starts with the configuration wizard
func NewWithWizard(cfg *config.Config) Model {
	// Initialize theme from config
	darkMode := cfg.IsDarkMode()
	ApplyTheme(darkMode)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	m := Model{
		state:         StateWizardWelcome,
		instances:     nil,
		spinner:       s,
		textInput:     newTextInput(cfg.DefaultSessionName),
		worktreeInput: newTextInput(constants.DefaultWorktreePlaceholder),
		config:        cfg,
		darkMode:      darkMode,
	}

	// Initialize wizard state
	m.initWizardState(cfg)

	return m
}

// initWizardState initializes wizard fields from a config
func (m *Model) initWizardState(cfg *config.Config) {
	// Initialize wizard inputs
	m.wizardPathInput = newTextInput("~/projects")
	m.wizardCredValue = newTextInput("~/.github_token")
	m.wizardSessionInput = newTextInput(constants.DefaultSessionName)
	m.wizardTimeoutInput = newTextInput("300")
	m.wizardLaunchInput = newTextInput("claude")
	m.wizardMaxDepthInput = newTextInput("3")

	// Initialize wizard state from config
	m.wizardSearchPaths = make([]string, len(cfg.SearchPaths))
	copy(m.wizardSearchPaths, cfg.SearchPaths)

	m.wizardCredentials = make([]auth.Credential, len(cfg.Auth.Credentials))
	copy(m.wizardCredentials, cfg.Auth.Credentials)

	m.wizardDarkMode = cfg.IsDarkMode()
	m.wizardPathWarnings = make(map[string]bool)

	// Set input values from config
	m.wizardSessionInput.SetValue(cfg.DefaultSessionName)
	m.wizardTimeoutInput.SetValue(fmt.Sprintf("%d", cfg.ContainerTimeout))
	m.wizardLaunchInput.SetValue(cfg.LaunchCommand)
	m.wizardMaxDepthInput.SetValue(fmt.Sprintf("%d", cfg.MaxDepth))

	m.wizardCredSource = auth.SourceFile
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
		m.warning = msg.authWarning
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
		// Store push warning for display (clear any previous warning)
		m.warning = msg.pushWarning
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

		// Combine warnings for display
		var warnings []string
		if msg.pushWarning != "" {
			warnings = append(warnings, msg.pushWarning)
		}
		if msg.labelWarning != "" {
			warnings = append(warnings, msg.labelWarning)
		}
		if len(warnings) > 0 {
			m.warning = strings.Join(warnings, "; ")
		}

		// Set up auto-start for after discovery completes
		m.pendingAutoStart = true
		m.autoStartWorktreePath = msg.worktreePath
		m.state = StateDiscovering
		return m, tea.Batch(m.spinner.Tick, m.discoverInstances())

	case wizardPathValidatedMsg:
		// Update path validation warnings
		if m.wizardPathWarnings == nil {
			m.wizardPathWarnings = make(map[string]bool)
		}
		m.wizardPathWarnings[msg.path] = msg.exists
		return m, nil

	case wizardConfigSavedMsg:
		// Config saved successfully, reload and go to dashboard
		newCfg, err := config.Load()
		if err != nil {
			m.state = StateError
			m.err = err
			m.errHint = "Press any key to go back"
			return m, nil
		}
		m.config = newCfg
		m.state = StateDiscovering
		return m, tea.Batch(m.spinner.Tick, m.discoverInstances())

	case wizardConfigErrorMsg:
		m.state = StateError
		m.err = msg.err
		m.errHint = "Press any key to go back"
		return m, nil
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
		return RenderDashboard(m.instancesStatus, m.cursor, m.width, m.warning)

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
		return RenderTmuxSelect(m.getInstanceName(), m.tmuxSessions, m.cursor, m.warning)

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

	case StateWizardWelcome:
		return RenderWizardWelcome(m.width)

	case StateWizardSearchPaths:
		return RenderWizardSearchPaths(m.wizardSearchPaths, m.wizardPathWarnings, m.wizardCursor, m.wizardPathInput, m.wizardEditMode, m.width)

	case StateWizardCredentials:
		return RenderWizardCredentials(m.wizardCredentials, m.wizardCredSource, m.wizardCredValue, m.wizardCursor, m.wizardEditMode, m.width)

	case StateWizardSettings:
		var activeInput textinput.Model
		switch m.wizardCursor {
		case 0:
			activeInput = m.wizardSessionInput
		case 1:
			activeInput = m.wizardTimeoutInput
		case 2:
			activeInput = m.wizardLaunchInput
		case 3:
			activeInput = m.wizardMaxDepthInput
		}
		return RenderWizardSettings(
			m.wizardSessionInput.Value(),
			m.wizardTimeoutInput.Value(),
			m.wizardLaunchInput.Value(),
			m.wizardMaxDepthInput.Value(),
			m.wizardDarkMode,
			m.wizardCursor,
			m.wizardEditMode,
			activeInput,
			m.width,
		)

	case StateWizardSummary:
		configPath, _ := getWizardConfigPath()
		// Show validated values (what will actually be saved)
		timeout := m.wizardTimeoutInput.Value()
		if t, err := strconv.Atoi(timeout); err != nil || t <= 0 {
			timeout = strconv.Itoa(constants.DefaultContainerTimeout)
		}
		maxDepth := m.wizardMaxDepthInput.Value()
		if d, err := strconv.Atoi(maxDepth); err != nil || d <= 0 {
			maxDepth = strconv.Itoa(constants.DefaultMaxDepth)
		}
		return RenderWizardSummary(
			m.wizardSearchPaths,
			m.wizardCredentials,
			m.wizardSessionInput.Value(),
			timeout,
			m.wizardLaunchInput.Value(),
			maxDepth,
			m.wizardDarkMode,
			configPath,
			m.width,
		)

	case StateWizardSaving:
		return RenderWizardSaving(m.spinner.View())
	}

	return ""
}
