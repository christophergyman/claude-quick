package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/christophergyman/claude-quick/internal/auth"
)

// handleWizardKey dispatches to the appropriate wizard step handler
func (m Model) handleWizardKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global wizard keys
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	}

	switch m.state {
	case StateWizardWelcome:
		return m.handleWizardWelcomeKey(msg)
	case StateWizardSearchPaths:
		return m.handleWizardSearchPathsKey(msg)
	case StateWizardCredentials:
		return m.handleWizardCredentialsKey(msg)
	case StateWizardSettings:
		return m.handleWizardSettingsKey(msg)
	case StateWizardSummary:
		return m.handleWizardSummaryKey(msg)
	}
	return m, nil
}

// handleWizardWelcomeKey handles keypresses on the welcome screen
func (m Model) handleWizardWelcomeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "n", " ":
		// Proceed to search paths
		m.state = StateWizardSearchPaths
		m.wizardCursor = 0
		return m, nil

	case "q", "esc":
		if m.wizardFromDashboard {
			m.state = StateDashboard
			return m, nil
		}
		return m, tea.Quit
	}
	return m, nil
}

// handleWizardSearchPathsKey handles keypresses on the search paths screen
func (m Model) handleWizardSearchPathsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.wizardEditMode {
		// Handle input mode
		switch msg.String() {
		case "enter":
			// Add the path
			path := m.wizardPathInput.Value()
			m.wizardEditMode = false
			m.wizardPathInput.Reset()
			if path != "" {
				m.wizardSearchPaths = append(m.wizardSearchPaths, path)
				// Validate the path
				return m, m.validateWizardPath(path)
			}
			return m, nil

		case "esc":
			m.wizardEditMode = false
			m.wizardPathInput.Reset()
			return m, nil

		default:
			var cmd tea.Cmd
			m.wizardPathInput, cmd = m.wizardPathInput.Update(msg)
			return m, cmd
		}
	}

	// Navigation mode
	switch msg.String() {
	case "up", "k":
		if m.wizardCursor > 0 {
			m.wizardCursor--
		}
		return m, nil

	case "down", "j":
		if m.wizardCursor < len(m.wizardSearchPaths)-1 {
			m.wizardCursor++
		}
		return m, nil

	case "a", "n":
		// Add new path
		m.wizardEditMode = true
		m.wizardPathInput.Reset()
		m.wizardPathInput.Focus()
		return m, textinput.Blink

	case "d", "x":
		// Delete selected path
		if len(m.wizardSearchPaths) > 0 && m.wizardCursor < len(m.wizardSearchPaths) {
			path := m.wizardSearchPaths[m.wizardCursor]
			m.wizardSearchPaths = append(m.wizardSearchPaths[:m.wizardCursor], m.wizardSearchPaths[m.wizardCursor+1:]...)
			delete(m.wizardPathWarnings, path)
			if m.wizardCursor >= len(m.wizardSearchPaths) && m.wizardCursor > 0 {
				m.wizardCursor--
			}
		}
		return m, nil

	case "enter", "tab":
		// Proceed to credentials
		m.state = StateWizardCredentials
		m.wizardCursor = 0
		m.wizardEditMode = false
		return m, nil

	case "backspace":
		// Go back to welcome
		m.state = StateWizardWelcome
		return m, nil

	case "q", "esc":
		if m.wizardFromDashboard {
			m.state = StateDashboard
			return m, nil
		}
		return m, tea.Quit
	}
	return m, nil
}

// handleWizardCredentialsKey handles keypresses on the credentials screen
func (m Model) handleWizardCredentialsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.wizardEditMode {
		// Handle credential input mode
		switch msg.String() {
		case "up", "k":
			// Cycle source type
			switch m.wizardCredSource {
			case auth.SourceEnv:
				m.wizardCredSource = auth.SourceFile
			case auth.SourceCommand:
				m.wizardCredSource = auth.SourceEnv
			}
			return m, nil

		case "down", "j":
			// Cycle source type
			switch m.wizardCredSource {
			case auth.SourceFile:
				m.wizardCredSource = auth.SourceEnv
			case auth.SourceEnv:
				m.wizardCredSource = auth.SourceCommand
			}
			return m, nil

		case "enter":
			// Add the credential
			value := m.wizardCredValue.Value()
			if value != "" {
				// Check if GITHUB_TOKEN already exists
				alreadyExists := false
				for _, c := range m.wizardCredentials {
					if c.Name == "GITHUB_TOKEN" {
						alreadyExists = true
						break
					}
				}
				if !alreadyExists {
					cred := auth.Credential{
						Name:   "GITHUB_TOKEN",
						Source: m.wizardCredSource,
						Value:  value,
					}
					m.wizardCredentials = append(m.wizardCredentials, cred)
				}
			}
			m.wizardEditMode = false
			m.wizardCredValue.Reset()
			return m, nil

		case "esc":
			m.wizardEditMode = false
			m.wizardCredValue.Reset()
			return m, nil

		default:
			var cmd tea.Cmd
			m.wizardCredValue, cmd = m.wizardCredValue.Update(msg)
			return m, cmd
		}
	}

	// Navigation mode
	switch msg.String() {
	case "up", "k":
		if m.wizardCursor > 0 {
			m.wizardCursor--
		}
		return m, nil

	case "down", "j":
		if m.wizardCursor < len(m.wizardCredentials)-1 {
			m.wizardCursor++
		}
		return m, nil

	case "a", "n":
		// Add new credential
		m.wizardEditMode = true
		m.wizardCredSource = auth.SourceFile
		m.wizardCredValue.Reset()
		m.wizardCredValue.Focus()
		return m, textinput.Blink

	case "d", "x":
		// Delete selected credential
		if len(m.wizardCredentials) > 0 && m.wizardCursor < len(m.wizardCredentials) {
			m.wizardCredentials = append(m.wizardCredentials[:m.wizardCursor], m.wizardCredentials[m.wizardCursor+1:]...)
			if m.wizardCursor >= len(m.wizardCredentials) && m.wizardCursor > 0 {
				m.wizardCursor--
			}
		}
		return m, nil

	case "s", "enter", "tab":
		// Skip/proceed to settings
		m.state = StateWizardSettings
		m.wizardCursor = 0
		m.wizardEditMode = false
		return m, nil

	case "backspace":
		// Go back to search paths
		m.state = StateWizardSearchPaths
		m.wizardCursor = 0
		return m, nil

	case "q", "esc":
		if m.wizardFromDashboard {
			m.state = StateDashboard
			return m, nil
		}
		return m, tea.Quit
	}
	return m, nil
}

// handleWizardSettingsKey handles keypresses on the settings screen
func (m Model) handleWizardSettingsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Settings: 0=session name, 1=timeout, 2=launch cmd, 3=max depth, 4=dark mode
	const settingsCount = 5

	if m.wizardEditMode {
		// Handle input mode for text fields
		switch msg.String() {
		case "enter":
			m.wizardEditMode = false
			return m, nil

		case "esc":
			m.wizardEditMode = false
			// Note: typed value persists in the input; user can re-edit to change
			return m, nil

		default:
			// Update the appropriate input
			var cmd tea.Cmd
			switch m.wizardCursor {
			case 0:
				m.wizardSessionInput, cmd = m.wizardSessionInput.Update(msg)
			case 1:
				m.wizardTimeoutInput, cmd = m.wizardTimeoutInput.Update(msg)
			case 2:
				m.wizardLaunchInput, cmd = m.wizardLaunchInput.Update(msg)
			case 3:
				m.wizardMaxDepthInput, cmd = m.wizardMaxDepthInput.Update(msg)
			}
			return m, cmd
		}
	}

	// Navigation mode
	switch msg.String() {
	case "up", "k":
		if m.wizardCursor > 0 {
			m.wizardCursor--
		}
		return m, nil

	case "down", "j":
		if m.wizardCursor < settingsCount-1 {
			m.wizardCursor++
		}
		return m, nil

	case "enter":
		// Edit selected field (except dark mode toggle)
		if m.wizardCursor == 4 {
			// Toggle dark mode
			m.wizardDarkMode = !m.wizardDarkMode
			return m, nil
		}
		m.wizardEditMode = true
		// Focus the appropriate input
		switch m.wizardCursor {
		case 0:
			m.wizardSessionInput.Focus()
		case 1:
			m.wizardTimeoutInput.Focus()
		case 2:
			m.wizardLaunchInput.Focus()
		case 3:
			m.wizardMaxDepthInput.Focus()
		}
		return m, textinput.Blink

	case "t":
		// Toggle dark mode regardless of cursor position
		m.wizardDarkMode = !m.wizardDarkMode
		return m, nil

	case "tab":
		// Proceed to summary
		m.state = StateWizardSummary
		m.wizardCursor = 0
		return m, nil

	case "backspace":
		// Go back to credentials
		m.state = StateWizardCredentials
		m.wizardCursor = 0
		return m, nil

	case "q", "esc":
		if m.wizardFromDashboard {
			m.state = StateDashboard
			return m, nil
		}
		return m, tea.Quit
	}
	return m, nil
}

// handleWizardSummaryKey handles keypresses on the summary screen
func (m Model) handleWizardSummaryKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "s":
		// Save configuration
		m.state = StateWizardSaving
		return m, tea.Batch(m.spinner.Tick, m.saveWizardConfig())

	case "backspace", "e":
		// Go back to edit (settings)
		m.state = StateWizardSettings
		m.wizardCursor = 0
		return m, nil

	case "q", "esc":
		if m.wizardFromDashboard {
			m.state = StateDashboard
			return m, nil
		}
		return m, tea.Quit
	}
	return m, nil
}
