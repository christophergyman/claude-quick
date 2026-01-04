package tui

import (
	"fmt"
	"os/exec"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/christophergyman/claude-quick/internal/config"
	"github.com/christophergyman/claude-quick/internal/devcontainer"
	"github.com/christophergyman/claude-quick/internal/tmux"
)

// State represents the current view state
type State int

const (
	StateDiscovering State = iota
	StateRefreshingStatus
	StateDashboard
	StateContainerStarting
	StateConfirmStop
	StateConfirmRestart
	StateContainerStopping
	StateContainerRestarting
	StateTmuxSelect
	StateNewSessionInput
	StateAttaching
	StateConfirmTmuxStop
	StateConfirmTmuxRestart
	StateTmuxStopping
	StateTmuxRestarting
	StateLoadingTmuxSessions
	StateError
	StateShowConfig
	StateNewWorktreeInput    // Text input for new worktree branch name
	StateCreatingWorktree    // Creating new worktree
	StateConfirmDeleteWorktree // Confirmation for deleting worktree
	StateDeletingWorktree    // Deleting worktree
)

// Model is the main Bubbletea model
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
	worktreeInput    textinput.Model // For new worktree branch name input
	err              error
	errHint          string
	width            int
	height           int
	config           *config.Config
	previousState    State
}

// Messages for async operations
type instancesDiscoveredMsg struct {
	instances []devcontainer.ContainerInstance
}
type instanceStatusRefreshedMsg struct {
	statuses []devcontainer.ContainerInstanceWithStatus
}
type containerStartedMsg struct{}
type containerErrorMsg struct{ err error }
type tmuxSessionsLoadedMsg struct{ sessions []string }
type tmuxSessionCreatedMsg struct{}
type containerStoppedMsg struct{}
type containerRestartedMsg struct{}
type tmuxSessionStoppedMsg struct{}
type tmuxSessionRestartedMsg struct{}
type tmuxDetachedMsg struct{}
type worktreeCreatedMsg struct {
	worktreePath string
}
type worktreeDeletedMsg struct{}

// New creates a new Model with discovered instances
func New(instances []devcontainer.ContainerInstance, cfg *config.Config) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	ti := textinput.New()
	ti.Placeholder = cfg.DefaultSessionName
	ti.CharLimit = 50
	ti.Width = 30

	wi := textinput.New()
	wi.Placeholder = "feature-branch"
	wi.CharLimit = 50
	wi.Width = 30

	return Model{
		state:         StateDashboard,
		instances:     instances,
		spinner:       s,
		textInput:     ti,
		worktreeInput: wi,
		config:        cfg,
	}
}

// NewWithDiscovery creates a Model that will discover instances asynchronously
func NewWithDiscovery(cfg *config.Config) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	ti := textinput.New()
	ti.Placeholder = cfg.DefaultSessionName
	ti.CharLimit = 50
	ti.Width = 30

	wi := textinput.New()
	wi.Placeholder = "feature-branch"
	wi.CharLimit = 50
	wi.Width = 30

	return Model{
		state:         StateDiscovering,
		instances:     nil,
		spinner:       s,
		textInput:     ti,
		worktreeInput: wi,
		config:        cfg,
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	if m.state == StateDiscovering {
		return tea.Batch(m.spinner.Tick, m.discoverInstances())
	}
	return nil
}

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
		m.state = StateDashboard
		return m, nil

	case tmuxDetachedMsg:
		// User detached from tmux, return to dashboard with status refresh
		m.state = StateRefreshingStatus
		m.selectedInstance = nil
		m.cursor = 0
		return m, tea.Batch(m.spinner.Tick, m.refreshInstanceStatus())

	case containerStartedMsg:
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
		// Worktree created, refresh instances and start the container for the new worktree
		m.state = StateDiscovering
		return m, tea.Batch(m.spinner.Tick, m.discoverInstances())

	case worktreeDeletedMsg:
		// Worktree deleted, refresh instances
		m.state = StateDiscovering
		m.selectedInstance = nil
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
	case "N", "esc":
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

// deleteWorktree removes the selected git worktree
func (m Model) deleteWorktree() tea.Cmd {
	return func() tea.Msg {
		if m.selectedInstance == nil {
			return containerErrorMsg{err: fmt.Errorf("no worktree selected")}
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
			return containerErrorMsg{err: nil}
		}
		worktreePath, err := devcontainer.CreateWorktree(m.selectedInstance.Path, branchName)
		if err != nil {
			return containerErrorMsg{err: err}
		}
		return worktreeCreatedMsg{worktreePath: worktreePath}
	}
}

// startContainer returns a command that starts the devcontainer
func (m Model) startContainer() tea.Cmd {
	return func() tea.Msg {
		if m.selectedInstance == nil {
			return containerErrorMsg{err: nil}
		}

		// Check if devcontainer CLI is available
		if err := devcontainer.CheckCLI(); err != nil {
			return containerErrorMsg{err: err}
		}

		// Start the container (path-based, each worktree has unique path)
		if err := devcontainer.Up(m.selectedInstance.Path); err != nil {
			return containerErrorMsg{err: err}
		}

		// Check if tmux is available in container
		if !devcontainer.HasTmux(m.selectedInstance.Path) {
			return containerErrorMsg{err: &tmuxNotFoundError{}}
		}

		return containerStartedMsg{}
	}
}

// stopContainer returns a command that stops the devcontainer
func (m Model) stopContainer() tea.Cmd {
	return func() tea.Msg {
		if m.selectedInstance == nil {
			return containerErrorMsg{err: nil}
		}
		if err := devcontainer.Stop(m.selectedInstance.Path); err != nil {
			return containerErrorMsg{err: err}
		}
		return containerStoppedMsg{}
	}
}

// restartContainer returns a command that restarts the devcontainer
func (m Model) restartContainer() tea.Cmd {
	return func() tea.Msg {
		if m.selectedInstance == nil {
			return containerErrorMsg{err: nil}
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
		if m.selectedInstance == nil || m.selectedSession == nil {
			return containerErrorMsg{err: nil}
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
		if m.selectedInstance == nil || m.selectedSession == nil {
			return containerErrorMsg{err: nil}
		}
		sessionName := m.selectedSession.Name
		// Kill existing session
		if err := devcontainer.KillTmuxSession(m.selectedInstance.Path, sessionName); err != nil {
			return containerErrorMsg{err: err}
		}
		// Create new session with same name
		if err := devcontainer.CreateTmuxSession(m.selectedInstance.Path, sessionName); err != nil {
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
		if err := devcontainer.CreateTmuxSession(m.selectedInstance.Path, name); err != nil {
			return containerErrorMsg{err: err}
		}
		return tmuxSessionCreatedMsg{}
	}
}

// attachToSession attaches to a tmux session using tea.ExecProcess
// This suspends the TUI, runs tmux as a subprocess, and returns to TUI on detach
func (m Model) attachToSession(sessionName string) (tea.Model, tea.Cmd) {
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
		instanceName := ""
		if m.selectedInstance != nil {
			instanceName = m.selectedInstance.DisplayName()
		}
		return RenderContainerStarting(instanceName, m.spinner.View())

	case StateConfirmStop:
		instanceName := ""
		if m.selectedInstance != nil {
			instanceName = m.selectedInstance.DisplayName()
		}
		return RenderConfirmDialog("stop", instanceName)

	case StateConfirmRestart:
		instanceName := ""
		if m.selectedInstance != nil {
			instanceName = m.selectedInstance.DisplayName()
		}
		return RenderConfirmDialog("restart", instanceName)

	case StateContainerStopping:
		instanceName := ""
		if m.selectedInstance != nil {
			instanceName = m.selectedInstance.DisplayName()
		}
		return RenderContainerOperation("Stopping", instanceName, m.spinner.View())

	case StateContainerRestarting:
		instanceName := ""
		if m.selectedInstance != nil {
			instanceName = m.selectedInstance.DisplayName()
		}
		return RenderContainerOperation("Restarting", instanceName, m.spinner.View())

	case StateConfirmTmuxStop:
		sessionName := ""
		if m.selectedSession != nil {
			sessionName = m.selectedSession.Name
		}
		return RenderTmuxConfirmDialog("stop", sessionName)

	case StateConfirmTmuxRestart:
		sessionName := ""
		if m.selectedSession != nil {
			sessionName = m.selectedSession.Name
		}
		return RenderTmuxConfirmDialog("restart", sessionName)

	case StateTmuxStopping:
		sessionName := ""
		if m.selectedSession != nil {
			sessionName = m.selectedSession.Name
		}
		return RenderTmuxOperation("Stopping", sessionName, m.spinner.View())

	case StateTmuxRestarting:
		sessionName := ""
		if m.selectedSession != nil {
			sessionName = m.selectedSession.Name
		}
		return RenderTmuxOperation("Restarting", sessionName, m.spinner.View())

	case StateLoadingTmuxSessions:
		instanceName := ""
		if m.selectedInstance != nil {
			instanceName = m.selectedInstance.DisplayName()
		}
		return RenderLoadingTmuxSessions(instanceName, m.spinner.View())

	case StateTmuxSelect:
		instanceName := ""
		if m.selectedInstance != nil {
			instanceName = m.selectedInstance.DisplayName()
		}
		return RenderTmuxSelect(instanceName, m.tmuxSessions, m.cursor)

	case StateNewSessionInput:
		instanceName := ""
		if m.selectedInstance != nil {
			instanceName = m.selectedInstance.DisplayName()
		}
		return RenderNewSessionInput(instanceName, m.textInput)

	case StateAttaching:
		instanceName := ""
		sessionName := m.textInput.Value()
		if m.selectedInstance != nil {
			instanceName = m.selectedInstance.DisplayName()
		}
		if m.cursor < len(m.tmuxSessions) {
			sessionName = m.tmuxSessions[m.cursor].Name
		}
		return RenderAttaching(instanceName, sessionName, m.spinner.View())

	case StateNewWorktreeInput:
		projectName := ""
		if m.selectedInstance != nil {
			projectName = m.selectedInstance.Name
		}
		return RenderNewWorktreeInput(projectName, m.worktreeInput)

	case StateCreatingWorktree:
		branchName := m.worktreeInput.Value()
		return RenderCreatingWorktree(branchName, m.spinner.View())

	case StateConfirmDeleteWorktree:
		branchName := ""
		if m.selectedInstance != nil && m.selectedInstance.Worktree != nil {
			branchName = m.selectedInstance.Worktree.Branch
		}
		return RenderConfirmDeleteWorktree(branchName)

	case StateDeletingWorktree:
		branchName := ""
		if m.selectedInstance != nil && m.selectedInstance.Worktree != nil {
			branchName = m.selectedInstance.Worktree.Branch
		}
		return RenderDeletingWorktree(branchName, m.spinner.View())

	case StateError:
		return RenderError(m.err, m.errHint)

	case StateShowConfig:
		return RenderConfigDisplay(m.config)
	}

	return ""
}

// tmuxNotFoundError indicates tmux is not available in the container
type tmuxNotFoundError struct{}

func (e *tmuxNotFoundError) Error() string {
	return "tmux not found in container. Please install tmux in your devcontainer."
}
