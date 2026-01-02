package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/chezu/quickvibe/internal/devcontainer"
	"github.com/chezu/quickvibe/internal/tmux"
)

// State represents the current view state
type State int

const (
	StateContainerSelect State = iota
	StateContainerStarting
	StateTmuxSelect
	StateNewSessionInput
	StateAttaching
	StateError
)

// Model is the main Bubbletea model
type Model struct {
	state           State
	projects        []devcontainer.Project
	selectedProject *devcontainer.Project
	tmuxSessions    []tmux.Session
	cursor          int
	spinner         spinner.Model
	textInput       textinput.Model
	err             error
	errHint         string
	width           int
	height          int
}

// Messages for async operations
type containerStartedMsg struct{}
type containerErrorMsg struct{ err error }
type tmuxSessionsLoadedMsg struct{ sessions []string }
type tmuxSessionCreatedMsg struct{}
type attachMsg struct{ sessionName string }

// New creates a new Model with discovered projects
func New(projects []devcontainer.Project) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	ti := textinput.New()
	ti.Placeholder = "session-name"
	ti.CharLimit = 50
	ti.Width = 30

	return Model{
		state:     StateContainerSelect,
		projects:  projects,
		spinner:   s,
		textInput: ti,
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
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

	case attachMsg:
		// This triggers the actual attachment - we exit the TUI
		return m, tea.Quit
	}

	// Update text input if in input state
	if m.state == StateNewSessionInput {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

// handleKeyPress processes keyboard input based on current state
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case StateContainerSelect:
		return m.handleContainerSelectKey(msg)
	case StateTmuxSelect:
		return m.handleTmuxSelectKey(msg)
	case StateNewSessionInput:
		return m.handleNewSessionInputKey(msg)
	case StateError:
		// Any key returns to container select
		m.state = StateContainerSelect
		m.err = nil
		return m, nil
	}
	return m, nil
}

func (m Model) handleContainerSelectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.projects)-1 {
			m.cursor++
		}

	case "enter":
		if len(m.projects) > 0 {
			m.selectedProject = &m.projects[m.cursor]
			m.state = StateContainerStarting
			return m, tea.Batch(
				m.spinner.Tick,
				m.startContainer(),
			)
		}
	}
	return m, nil
}

func (m Model) handleTmuxSelectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	totalOptions := TotalTmuxOptions(m.tmuxSessions)

	switch msg.String() {
	case "q", "esc":
		// Go back to container select
		m.state = StateContainerSelect
		m.cursor = 0
		m.selectedProject = nil
		return m, nil

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
		if name != "" {
			m.state = StateAttaching
			return m, tea.Batch(
				m.spinner.Tick,
				m.createTmuxSession(name),
			)
		}
	}

	// Pass other keys to text input
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// startContainer returns a command that starts the devcontainer
func (m Model) startContainer() tea.Cmd {
	return func() tea.Msg {
		if m.selectedProject == nil {
			return containerErrorMsg{err: nil}
		}

		// Check if devcontainer CLI is available
		if err := devcontainer.CheckCLI(); err != nil {
			return containerErrorMsg{err: err}
		}

		// Start the container
		if err := devcontainer.Up(m.selectedProject.Path); err != nil {
			return containerErrorMsg{err: err}
		}

		// Check if tmux is available in container
		if !devcontainer.HasTmux(m.selectedProject.Path) {
			return containerErrorMsg{err: &tmuxNotFoundError{}}
		}

		return containerStartedMsg{}
	}
}

// handleContainerStarted is called when container has started
func (m Model) handleContainerStarted() (tea.Model, tea.Cmd) {
	// Load tmux sessions
	return m, func() tea.Msg {
		sessions, err := devcontainer.ListTmuxSessions(m.selectedProject.Path)
		if err != nil {
			return containerErrorMsg{err: err}
		}
		return tmuxSessionsLoadedMsg{sessions: sessions}
	}
}

// createTmuxSession creates a new tmux session in the container
func (m Model) createTmuxSession(name string) tea.Cmd {
	return func() tea.Msg {
		if err := devcontainer.CreateTmuxSession(m.selectedProject.Path, name); err != nil {
			return containerErrorMsg{err: err}
		}
		return tmuxSessionCreatedMsg{}
	}
}

// attachToSession prepares to attach to a tmux session
func (m Model) attachToSession(sessionName string) (tea.Model, tea.Cmd) {
	m.state = StateAttaching
	// Store the session name for use after quit
	return m, tea.Sequence(
		tea.Batch(m.spinner.Tick),
		func() tea.Msg {
			return attachMsg{sessionName: sessionName}
		},
	)
}

// View implements tea.Model
func (m Model) View() string {
	switch m.state {
	case StateContainerSelect:
		return RenderContainerSelect(m.projects, m.cursor, m.width)

	case StateContainerStarting:
		projectName := ""
		if m.selectedProject != nil {
			projectName = m.selectedProject.Name
		}
		return RenderContainerStarting(projectName, m.spinner.View())

	case StateTmuxSelect:
		projectName := ""
		if m.selectedProject != nil {
			projectName = m.selectedProject.Name
		}
		return RenderTmuxSelect(projectName, m.tmuxSessions, m.cursor)

	case StateNewSessionInput:
		projectName := ""
		if m.selectedProject != nil {
			projectName = m.selectedProject.Name
		}
		return RenderNewSessionInput(projectName, m.textInput)

	case StateAttaching:
		projectName := ""
		sessionName := m.textInput.Value()
		if m.selectedProject != nil {
			projectName = m.selectedProject.Name
		}
		if m.cursor < len(m.tmuxSessions) {
			sessionName = m.tmuxSessions[m.cursor].Name
		}
		return RenderAttaching(projectName, sessionName, m.spinner.View())

	case StateError:
		return RenderError(m.err, m.errHint)
	}

	return ""
}

// GetAttachInfo returns the info needed to attach after the TUI exits
func (m Model) GetAttachInfo() (projectPath, sessionName string, shouldAttach bool) {
	if m.state != StateAttaching || m.selectedProject == nil {
		return "", "", false
	}

	sessionName = m.textInput.Value()
	if m.cursor < len(m.tmuxSessions) {
		sessionName = m.tmuxSessions[m.cursor].Name
	}

	return m.selectedProject.Path, sessionName, true
}

// tmuxNotFoundError indicates tmux is not available in the container
type tmuxNotFoundError struct{}

func (e *tmuxNotFoundError) Error() string {
	return "tmux not found in container. Please install tmux in your devcontainer."
}
