package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/christophergyman/claude-quick/internal/auth"
	"github.com/christophergyman/claude-quick/internal/config"
	"github.com/christophergyman/claude-quick/internal/constants"
)

// ============================================================================
// wizard.go render tests
// ============================================================================

func TestRenderWizardProgress(t *testing.T) {
	tests := []struct {
		name        string
		currentStep int
		contains    []string
	}{
		{
			name:        "welcome step",
			currentStep: WizardStepWelcome,
			contains:    []string{"Step 1 of 5", "Welcome"},
		},
		{
			name:        "search paths step",
			currentStep: WizardStepSearchPaths,
			contains:    []string{"Step 2 of 5", "Paths"},
		},
		{
			name:        "credentials step",
			currentStep: WizardStepCredentials,
			contains:    []string{"Step 3 of 5", "Credentials"},
		},
		{
			name:        "settings step",
			currentStep: WizardStepSettings,
			contains:    []string{"Step 4 of 5", "Settings"},
		},
		{
			name:        "summary step",
			currentStep: WizardStepSummary,
			contains:    []string{"Step 5 of 5", "Summary"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderWizardProgress(tt.currentStep, 60)

			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("renderWizardProgress(%d) should contain %q", tt.currentStep, expected)
				}
			}
		})
	}
}

func TestRenderWizardWelcome(t *testing.T) {
	result := RenderWizardWelcome(65)

	expectedContents := []string{
		"claude-quick",
		"Configuration Wizard",
		"Welcome",
		"search paths",
		"credentials",
		"settings",
		"enter",
		"cancel",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(strings.ToLower(result), strings.ToLower(expected)) {
			t.Errorf("RenderWizardWelcome should contain %q", expected)
		}
	}
}

func TestRenderWizardSearchPaths(t *testing.T) {
	tests := []struct {
		name     string
		paths    []string
		warnings map[string]bool
		cursor   int
		editMode bool
		contains []string
	}{
		{
			name:     "empty paths",
			paths:    []string{},
			warnings: nil,
			cursor:   0,
			editMode: false,
			contains: []string{"No paths configured", "add"},
		},
		{
			name:     "with paths",
			paths:    []string{"~/projects", "/home/user/work"},
			warnings: map[string]bool{"~/projects": true, "/home/user/work": false},
			cursor:   0,
			editMode: false,
			contains: []string{"~/projects", "/home/user/work", "not found"},
		},
		{
			name:     "edit mode",
			paths:    []string{},
			warnings: nil,
			cursor:   0,
			editMode: true,
			contains: []string{"Enter path", "enter add", "esc cancel"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ti := textinput.New()
			result := RenderWizardSearchPaths(tt.paths, tt.warnings, tt.cursor, ti, tt.editMode, 65)

			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("RenderWizardSearchPaths should contain %q", expected)
				}
			}
		})
	}
}

func TestRenderWizardCredentials(t *testing.T) {
	tests := []struct {
		name        string
		credentials []auth.Credential
		sourceType  auth.SourceType
		cursor      int
		editMode    bool
		contains    []string
	}{
		{
			name:        "no credentials",
			credentials: []auth.Credential{},
			sourceType:  auth.SourceFile,
			cursor:      0,
			editMode:    false,
			contains:    []string{"No credentials configured", "optional"},
		},
		{
			name: "with credential",
			credentials: []auth.Credential{
				{Name: "GITHUB_TOKEN", Source: auth.SourceFile, Value: "~/.github_token"},
			},
			sourceType: auth.SourceFile,
			cursor:     0,
			editMode:   false,
			contains:   []string{"GITHUB_TOKEN", "file", "~/.github_token"},
		},
		{
			name:        "edit mode",
			credentials: []auth.Credential{},
			sourceType:  auth.SourceEnv,
			cursor:      0,
			editMode:    true,
			contains:    []string{"Select source type", "file", "env", "command"},
		},
		{
			name: "edit mode with existing token",
			credentials: []auth.Credential{
				{Name: "GITHUB_TOKEN", Source: auth.SourceFile, Value: "~/.token"},
			},
			sourceType: auth.SourceFile,
			cursor:     0,
			editMode:   true,
			contains:   []string{"already configured", "esc to cancel"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ti := textinput.New()
			result := RenderWizardCredentials(tt.credentials, tt.sourceType, ti, tt.cursor, tt.editMode, 65)

			for _, expected := range tt.contains {
				if !strings.Contains(strings.ToLower(result), strings.ToLower(expected)) {
					t.Errorf("RenderWizardCredentials should contain %q", expected)
				}
			}
		})
	}
}

func TestRenderWizardSettings(t *testing.T) {
	tests := []struct {
		name     string
		cursor   int
		editMode bool
		contains []string
	}{
		{
			name:     "all settings visible",
			cursor:   0,
			editMode: false,
			contains: []string{"session name", "timeout", "Launch command", "depth", "Dark mode"},
		},
		{
			name:     "edit mode",
			cursor:   0,
			editMode: true,
			contains: []string{"enter confirm", "esc cancel"},
		},
		{
			name:     "dark mode enabled",
			cursor:   4,
			editMode: false,
			contains: []string{"enabled"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ti := textinput.New()
			ti.SetValue("main")
			result := RenderWizardSettings("main", "300", "claude", "3", true, tt.cursor, tt.editMode, ti, 65)

			for _, expected := range tt.contains {
				if !strings.Contains(strings.ToLower(result), strings.ToLower(expected)) {
					t.Errorf("RenderWizardSettings should contain %q", expected)
				}
			}
		})
	}
}

func TestRenderWizardSummary(t *testing.T) {
	paths := []string{"~/projects", "/work"}
	creds := []auth.Credential{
		{Name: "GITHUB_TOKEN", Source: auth.SourceFile, Value: "~/.token"},
	}

	result := RenderWizardSummary(paths, creds, "main", "300", "claude", "3", true, "/path/to/config.yaml", 65)

	expectedContents := []string{
		"SEARCH PATHS",
		"~/projects",
		"/work",
		"CREDENTIALS",
		"GITHUB_TOKEN",
		"SETTINGS",
		"Default Session",
		"main",
		"300s",
		"claude",
		"Config will be saved to",
		"enter save",
		"backspace edit",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(result, expected) {
			t.Errorf("RenderWizardSummary should contain %q", expected)
		}
	}
}

func TestRenderWizardSummary_EmptyPaths(t *testing.T) {
	result := RenderWizardSummary([]string{}, []auth.Credential{}, "main", "300", "claude", "3", true, "/config.yaml", 65)

	if !strings.Contains(result, "(none)") {
		t.Error("RenderWizardSummary with empty paths should show '(none)'")
	}
}

func TestRenderWizardSummary_EmptyLaunchCommand(t *testing.T) {
	result := RenderWizardSummary([]string{"~/projects"}, []auth.Credential{}, "main", "300", "", "3", true, "/config.yaml", 65)

	if !strings.Contains(result, "(none)") {
		t.Error("RenderWizardSummary with empty launch command should show '(none)'")
	}
}

func TestRenderWizardSettings_InvalidNumericValues(t *testing.T) {
	ti := textinput.New()
	// Test with invalid timeout and maxDepth
	result := RenderWizardSettings("main", "abc", "claude", "xyz", true, 0, false, ti, 65)

	if !strings.Contains(result, "invalid") {
		t.Error("RenderWizardSettings should show 'invalid' warning for non-numeric values")
	}
	if !strings.Contains(result, "will use default") {
		t.Error("RenderWizardSettings should indicate defaults will be used for invalid values")
	}
}

func TestRenderWizardSettings_ValidNumericValues(t *testing.T) {
	ti := textinput.New()
	// Test with valid values
	result := RenderWizardSettings("main", "300", "claude", "3", true, 0, false, ti, 65)

	if strings.Contains(result, "invalid") {
		t.Error("RenderWizardSettings should not show 'invalid' warning for valid numeric values")
	}
}

func TestRenderWizardSaving(t *testing.T) {
	result := RenderWizardSaving("⠋")

	if !strings.Contains(result, "Saving") {
		t.Error("RenderWizardSaving should contain 'Saving'")
	}
	if !strings.Contains(result, "⠋") {
		t.Error("RenderWizardSaving should contain spinner")
	}
}

// ============================================================================
// wizard_handlers.go tests
// ============================================================================

func TestHandleWizardWelcomeKey(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		fromDashboard bool
		expectedState State
		shouldQuit    bool
	}{
		{
			name:          "enter proceeds to search paths",
			key:           "enter",
			fromDashboard: false,
			expectedState: StateWizardSearchPaths,
		},
		{
			name:          "space proceeds to search paths",
			key:           " ",
			fromDashboard: false,
			expectedState: StateWizardSearchPaths,
		},
		{
			name:          "n proceeds to search paths",
			key:           "n",
			fromDashboard: false,
			expectedState: StateWizardSearchPaths,
		},
		{
			name:          "esc from first run quits",
			key:           "esc",
			fromDashboard: false,
			shouldQuit:    true,
		},
		{
			name:          "q from first run quits",
			key:           "q",
			fromDashboard: false,
			shouldQuit:    true,
		},
		{
			name:          "esc from dashboard returns to dashboard",
			key:           "esc",
			fromDashboard: true,
			expectedState: StateDashboard,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				state:               StateWizardWelcome,
				wizardFromDashboard: tt.fromDashboard,
			}

			newModel, cmd := m.handleWizardWelcomeKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})
			model := newModel.(Model)

			if tt.shouldQuit {
				if cmd == nil {
					t.Error("expected quit command")
				}
			} else {
				if model.state != tt.expectedState {
					t.Errorf("state = %v, want %v", model.state, tt.expectedState)
				}
			}
		})
	}
}

func TestHandleWizardSearchPathsKey_Navigation(t *testing.T) {
	m := Model{
		state:             StateWizardSearchPaths,
		wizardSearchPaths: []string{"path1", "path2", "path3"},
		wizardCursor:      1,
		wizardEditMode:    false,
	}

	// Test up navigation
	newModel, _ := m.handleWizardSearchPathsKey(tea.KeyMsg{Type: tea.KeyUp})
	model := newModel.(Model)
	if model.wizardCursor != 0 {
		t.Errorf("up: cursor = %d, want 0", model.wizardCursor)
	}

	// Test down navigation
	m.wizardCursor = 1
	newModel, _ = m.handleWizardSearchPathsKey(tea.KeyMsg{Type: tea.KeyDown})
	model = newModel.(Model)
	if model.wizardCursor != 2 {
		t.Errorf("down: cursor = %d, want 2", model.wizardCursor)
	}

	// Test cursor bounds (at bottom)
	m.wizardCursor = 2
	newModel, _ = m.handleWizardSearchPathsKey(tea.KeyMsg{Type: tea.KeyDown})
	model = newModel.(Model)
	if model.wizardCursor != 2 {
		t.Errorf("down at bottom: cursor = %d, want 2", model.wizardCursor)
	}
}

func TestHandleWizardSearchPathsKey_AddPath(t *testing.T) {
	m := Model{
		state:             StateWizardSearchPaths,
		wizardSearchPaths: []string{},
		wizardEditMode:    false,
		wizardPathInput:   textinput.New(),
	}

	// Press 'a' to enter add mode
	newModel, cmd := m.handleWizardSearchPathsKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	model := newModel.(Model)

	if !model.wizardEditMode {
		t.Error("pressing 'a' should enter edit mode")
	}
	if cmd == nil {
		t.Error("should return textinput.Blink command")
	}
}

func TestHandleWizardSearchPathsKey_AddPathConfirm(t *testing.T) {
	ti := textinput.New()
	ti.SetValue("~/new-path")

	m := Model{
		state:              StateWizardSearchPaths,
		wizardSearchPaths:  []string{},
		wizardEditMode:     true,
		wizardPathInput:    ti,
		wizardPathWarnings: make(map[string]bool),
	}

	// Press enter to confirm
	newModel, cmd := m.handleWizardSearchPathsKey(tea.KeyMsg{Type: tea.KeyEnter})
	model := newModel.(Model)

	if model.wizardEditMode {
		t.Error("should exit edit mode after adding")
	}
	if len(model.wizardSearchPaths) != 1 {
		t.Errorf("paths = %d, want 1", len(model.wizardSearchPaths))
	}
	if model.wizardSearchPaths[0] != "~/new-path" {
		t.Errorf("path = %q, want ~/new-path", model.wizardSearchPaths[0])
	}
	if cmd == nil {
		t.Error("should return validation command")
	}
}

func TestHandleWizardSearchPathsKey_DeletePath(t *testing.T) {
	m := Model{
		state:              StateWizardSearchPaths,
		wizardSearchPaths:  []string{"path1", "path2", "path3"},
		wizardCursor:       1,
		wizardEditMode:     false,
		wizardPathWarnings: map[string]bool{"path1": true, "path2": true, "path3": true},
	}

	// Press 'd' to delete
	newModel, _ := m.handleWizardSearchPathsKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	model := newModel.(Model)

	if len(model.wizardSearchPaths) != 2 {
		t.Errorf("paths = %d, want 2", len(model.wizardSearchPaths))
	}
	if model.wizardSearchPaths[0] != "path1" || model.wizardSearchPaths[1] != "path3" {
		t.Errorf("wrong paths after delete: %v", model.wizardSearchPaths)
	}
	if _, exists := model.wizardPathWarnings["path2"]; exists {
		t.Error("deleted path warning should be removed")
	}
}

func TestHandleWizardSearchPathsKey_Proceed(t *testing.T) {
	m := Model{
		state:             StateWizardSearchPaths,
		wizardSearchPaths: []string{"path1"},
		wizardEditMode:    false,
	}

	// Press enter to proceed
	newModel, _ := m.handleWizardSearchPathsKey(tea.KeyMsg{Type: tea.KeyEnter})
	model := newModel.(Model)

	if model.state != StateWizardCredentials {
		t.Errorf("state = %v, want StateWizardCredentials", model.state)
	}
	if model.wizardCursor != 0 {
		t.Error("cursor should reset to 0")
	}
}

func TestHandleWizardCredentialsKey_AddCredential(t *testing.T) {
	ti := textinput.New()
	ti.SetValue("~/.github_token")

	m := Model{
		state:             StateWizardCredentials,
		wizardCredentials: []auth.Credential{},
		wizardCredSource:  auth.SourceFile,
		wizardCredValue:   ti,
		wizardEditMode:    true,
	}

	// Press enter to add
	newModel, _ := m.handleWizardCredentialsKey(tea.KeyMsg{Type: tea.KeyEnter})
	model := newModel.(Model)

	if len(model.wizardCredentials) != 1 {
		t.Errorf("credentials = %d, want 1", len(model.wizardCredentials))
	}
	if model.wizardCredentials[0].Name != "GITHUB_TOKEN" {
		t.Error("credential name should be GITHUB_TOKEN")
	}
	if model.wizardCredentials[0].Source != auth.SourceFile {
		t.Error("credential source should be file")
	}
}

func TestHandleWizardCredentialsKey_PreventDuplicate(t *testing.T) {
	ti := textinput.New()
	ti.SetValue("~/.new_token")

	m := Model{
		state: StateWizardCredentials,
		wizardCredentials: []auth.Credential{
			{Name: "GITHUB_TOKEN", Source: auth.SourceFile, Value: "~/.existing"},
		},
		wizardCredSource: auth.SourceEnv,
		wizardCredValue:  ti,
		wizardEditMode:   true,
	}

	// Press enter to try to add duplicate
	newModel, _ := m.handleWizardCredentialsKey(tea.KeyMsg{Type: tea.KeyEnter})
	model := newModel.(Model)

	if len(model.wizardCredentials) != 1 {
		t.Errorf("credentials = %d, want 1 (duplicate should be prevented)", len(model.wizardCredentials))
	}
	if model.wizardCredentials[0].Value != "~/.existing" {
		t.Error("original credential should remain unchanged")
	}
}

func TestHandleWizardCredentialsKey_CycleSource(t *testing.T) {
	m := Model{
		state:            StateWizardCredentials,
		wizardCredSource: auth.SourceFile,
		wizardCredValue:  textinput.New(),
		wizardEditMode:   true,
	}

	// Press down to cycle to env
	newModel, _ := m.handleWizardCredentialsKey(tea.KeyMsg{Type: tea.KeyDown})
	model := newModel.(Model)
	if model.wizardCredSource != auth.SourceEnv {
		t.Errorf("source = %v, want env", model.wizardCredSource)
	}

	// Press down to cycle to command
	newModel, _ = model.handleWizardCredentialsKey(tea.KeyMsg{Type: tea.KeyDown})
	model = newModel.(Model)
	if model.wizardCredSource != auth.SourceCommand {
		t.Errorf("source = %v, want command", model.wizardCredSource)
	}

	// Press up to cycle back to env
	newModel, _ = model.handleWizardCredentialsKey(tea.KeyMsg{Type: tea.KeyUp})
	model = newModel.(Model)
	if model.wizardCredSource != auth.SourceEnv {
		t.Errorf("source = %v, want env", model.wizardCredSource)
	}
}

func TestHandleWizardSettingsKey_Navigation(t *testing.T) {
	m := Model{
		state:              StateWizardSettings,
		wizardCursor:       0,
		wizardEditMode:     false,
		wizardSessionInput: textinput.New(),
	}

	// Navigate down
	newModel, _ := m.handleWizardSettingsKey(tea.KeyMsg{Type: tea.KeyDown})
	model := newModel.(Model)
	if model.wizardCursor != 1 {
		t.Errorf("cursor = %d, want 1", model.wizardCursor)
	}

	// Navigate to dark mode (index 4)
	m.wizardCursor = 4
	newModel, _ = m.handleWizardSettingsKey(tea.KeyMsg{Type: tea.KeyDown})
	model = newModel.(Model)
	if model.wizardCursor != 4 {
		t.Error("cursor should not go past 4")
	}
}

func TestHandleWizardSettingsKey_ToggleDarkMode(t *testing.T) {
	m := Model{
		state:          StateWizardSettings,
		wizardCursor:   4,
		wizardDarkMode: true,
		wizardEditMode: false,
	}

	// Press enter on dark mode to toggle
	newModel, _ := m.handleWizardSettingsKey(tea.KeyMsg{Type: tea.KeyEnter})
	model := newModel.(Model)
	if model.wizardDarkMode {
		t.Error("dark mode should be toggled to false")
	}

	// Press 't' to toggle from any cursor position
	m.wizardCursor = 0
	m.wizardDarkMode = false
	newModel, _ = m.handleWizardSettingsKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	model = newModel.(Model)
	if !model.wizardDarkMode {
		t.Error("dark mode should be toggled to true")
	}
}

func TestHandleWizardSettingsKey_EditField(t *testing.T) {
	ti := textinput.New()
	m := Model{
		state:              StateWizardSettings,
		wizardCursor:       0,
		wizardEditMode:     false,
		wizardSessionInput: ti,
	}

	// Press enter to edit
	newModel, cmd := m.handleWizardSettingsKey(tea.KeyMsg{Type: tea.KeyEnter})
	model := newModel.(Model)

	if !model.wizardEditMode {
		t.Error("should enter edit mode")
	}
	if cmd == nil {
		t.Error("should return textinput.Blink command")
	}
}

func TestHandleWizardSummaryKey_Save(t *testing.T) {
	m := Model{
		state:               StateWizardSummary,
		wizardSearchPaths:   []string{"~/projects"},
		wizardCredentials:   []auth.Credential{},
		wizardSessionInput:  textinput.New(),
		wizardTimeoutInput:  textinput.New(),
		wizardLaunchInput:   textinput.New(),
		wizardMaxDepthInput: textinput.New(),
	}

	// Press enter to save
	newModel, cmd := m.handleWizardSummaryKey(tea.KeyMsg{Type: tea.KeyEnter})
	model := newModel.(Model)

	if model.state != StateWizardSaving {
		t.Errorf("state = %v, want StateWizardSaving", model.state)
	}
	if cmd == nil {
		t.Error("should return save command")
	}
}

func TestHandleWizardSummaryKey_GoBack(t *testing.T) {
	m := Model{
		state: StateWizardSummary,
	}

	// Press backspace to go back
	newModel, _ := m.handleWizardSummaryKey(tea.KeyMsg{Type: tea.KeyBackspace})
	model := newModel.(Model)

	if model.state != StateWizardSettings {
		t.Errorf("state = %v, want StateWizardSettings", model.state)
	}
}

// ============================================================================
// model.go wizard tests
// ============================================================================

func TestNewWithWizard(t *testing.T) {
	cfg := config.DefaultConfig()
	model := NewWithWizard(cfg)

	if model.state != StateWizardWelcome {
		t.Errorf("state = %v, want StateWizardWelcome", model.state)
	}
	if model.wizardPathWarnings == nil {
		t.Error("wizardPathWarnings should be initialized")
	}
}

func TestInitWizardState(t *testing.T) {
	cfg := &config.Config{
		SearchPaths:        []string{"~/projects", "/work"},
		MaxDepth:           5,
		DefaultSessionName: "dev",
		ContainerTimeout:   600,
		LaunchCommand:      "npm start",
		Auth: auth.Config{
			Credentials: []auth.Credential{
				{Name: "GITHUB_TOKEN", Source: auth.SourceFile, Value: "~/.token"},
			},
		},
	}

	m := Model{}
	m.initWizardState(cfg)

	if len(m.wizardSearchPaths) != 2 {
		t.Errorf("wizardSearchPaths = %d, want 2", len(m.wizardSearchPaths))
	}
	if len(m.wizardCredentials) != 1 {
		t.Errorf("wizardCredentials = %d, want 1", len(m.wizardCredentials))
	}
	if m.wizardSessionInput.Value() != "dev" {
		t.Errorf("session input = %q, want dev", m.wizardSessionInput.Value())
	}
	if m.wizardTimeoutInput.Value() != "600" {
		t.Errorf("timeout input = %q, want 600", m.wizardTimeoutInput.Value())
	}
	if m.wizardLaunchInput.Value() != "npm start" {
		t.Errorf("launch input = %q, want 'npm start'", m.wizardLaunchInput.Value())
	}
	if m.wizardMaxDepthInput.Value() != "5" {
		t.Errorf("max depth input = %q, want 5", m.wizardMaxDepthInput.Value())
	}
}

func TestInitWizardState_CopiesSlices(t *testing.T) {
	cfg := &config.Config{
		SearchPaths: []string{"~/original"},
		Auth: auth.Config{
			Credentials: []auth.Credential{
				{Name: "TOKEN", Source: auth.SourceFile, Value: "~/.token"},
			},
		},
	}

	m := Model{}
	m.initWizardState(cfg)

	// Modify wizard state
	m.wizardSearchPaths[0] = "~/modified"
	m.wizardCredentials[0].Value = "~/.modified"

	// Original should be unchanged
	if cfg.SearchPaths[0] != "~/original" {
		t.Error("original config paths should not be modified")
	}
	if cfg.Auth.Credentials[0].Value != "~/.token" {
		t.Error("original config credentials should not be modified")
	}
}

// ============================================================================
// commands.go wizard tests
// ============================================================================

func TestBuildWizardConfig(t *testing.T) {
	sessionInput := textinput.New()
	sessionInput.SetValue("dev")

	timeoutInput := textinput.New()
	timeoutInput.SetValue("600")

	launchInput := textinput.New()
	launchInput.SetValue("npm start")

	maxDepthInput := textinput.New()
	maxDepthInput.SetValue("5")

	m := Model{
		wizardSearchPaths:   []string{"~/projects", "/work"},
		wizardCredentials:   []auth.Credential{{Name: "GITHUB_TOKEN", Source: auth.SourceFile, Value: "~/.token"}},
		wizardSessionInput:  sessionInput,
		wizardTimeoutInput:  timeoutInput,
		wizardLaunchInput:   launchInput,
		wizardMaxDepthInput: maxDepthInput,
		wizardDarkMode:      true,
	}

	cfg := m.buildWizardConfig()

	if len(cfg.SearchPaths) != 2 {
		t.Errorf("SearchPaths = %d, want 2", len(cfg.SearchPaths))
	}
	if cfg.DefaultSessionName != "dev" {
		t.Errorf("DefaultSessionName = %q, want dev", cfg.DefaultSessionName)
	}
	if cfg.ContainerTimeout != 600 {
		t.Errorf("ContainerTimeout = %d, want 600", cfg.ContainerTimeout)
	}
	if cfg.LaunchCommand != "npm start" {
		t.Errorf("LaunchCommand = %q, want 'npm start'", cfg.LaunchCommand)
	}
	if cfg.MaxDepth != 5 {
		t.Errorf("MaxDepth = %d, want 5", cfg.MaxDepth)
	}
	if cfg.DarkMode == nil || !*cfg.DarkMode {
		t.Error("DarkMode should be true")
	}
	if len(cfg.Auth.Credentials) != 1 {
		t.Errorf("Credentials = %d, want 1", len(cfg.Auth.Credentials))
	}
}

func TestBuildWizardConfig_InvalidInputs(t *testing.T) {
	sessionInput := textinput.New()
	sessionInput.SetValue("") // empty should use default

	timeoutInput := textinput.New()
	timeoutInput.SetValue("invalid") // invalid should use default

	launchInput := textinput.New()
	launchInput.SetValue("")

	maxDepthInput := textinput.New()
	maxDepthInput.SetValue("-1") // invalid should use default

	m := Model{
		wizardSearchPaths:   []string{},
		wizardSessionInput:  sessionInput,
		wizardTimeoutInput:  timeoutInput,
		wizardLaunchInput:   launchInput,
		wizardMaxDepthInput: maxDepthInput,
	}

	cfg := m.buildWizardConfig()

	if cfg.DefaultSessionName != constants.DefaultSessionName {
		t.Errorf("DefaultSessionName = %q, want default %q", cfg.DefaultSessionName, constants.DefaultSessionName)
	}
	if cfg.ContainerTimeout != constants.DefaultContainerTimeout {
		t.Errorf("ContainerTimeout = %d, want default %d", cfg.ContainerTimeout, constants.DefaultContainerTimeout)
	}
	if cfg.MaxDepth != constants.DefaultMaxDepth {
		t.Errorf("MaxDepth = %d, want default %d", cfg.MaxDepth, constants.DefaultMaxDepth)
	}
}

// ============================================================================
// config.go wizard tests
// ============================================================================

func TestConfigExists(t *testing.T) {
	// This tests the function exists and returns a boolean
	// The actual result depends on whether a config file exists
	exists := config.ConfigExists()
	t.Logf("ConfigExists() = %v", exists)
}

func TestGetExecutableDirConfigPath(t *testing.T) {
	path, err := config.GetExecutableDirConfigPath()
	if err != nil {
		t.Logf("GetExecutableDirConfigPath() returned error (may be expected in test env): %v", err)
		return
	}

	if path == "" {
		t.Error("GetExecutableDirConfigPath() returned empty path")
	}

	// Should end with claude-quick.yaml
	if filepath.Base(path) != "claude-quick.yaml" {
		t.Errorf("path = %q, should end with claude-quick.yaml", path)
	}
}

func TestConfigSave(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "wizard-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "claude-quick.yaml")

	cfg := &config.Config{
		SearchPaths:        []string{"~/projects"},
		MaxDepth:           3,
		DefaultSessionName: "main",
		ContainerTimeout:   300,
		LaunchCommand:      "claude",
	}

	err = config.Save(cfg, configPath)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}

	// Read and verify contents
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "~/projects") {
		t.Error("config should contain search path")
	}
	if !strings.Contains(content, "max_depth: 3") {
		t.Error("config should contain max_depth")
	}
}

func TestConfigSave_CreatesDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wizard-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Path with non-existent subdirectory
	configPath := filepath.Join(tmpDir, "subdir", "claude-quick.yaml")

	cfg := config.DefaultConfig()

	err = config.Save(cfg, configPath)
	if err != nil {
		t.Fatalf("Save() should create parent directory: %v", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}
}

// ============================================================================
// Wizard state constants tests
// ============================================================================

func TestWizardStepConstants(t *testing.T) {
	if WizardStepWelcome != 0 {
		t.Error("WizardStepWelcome should be 0")
	}
	if WizardStepSearchPaths != 1 {
		t.Error("WizardStepSearchPaths should be 1")
	}
	if WizardStepCredentials != 2 {
		t.Error("WizardStepCredentials should be 2")
	}
	if WizardStepSettings != 3 {
		t.Error("WizardStepSettings should be 3")
	}
	if WizardStepSummary != 4 {
		t.Error("WizardStepSummary should be 4")
	}
	if WizardStepCount != 5 {
		t.Error("WizardStepCount should be 5")
	}
}

func TestWizardStepNames(t *testing.T) {
	if len(wizardStepNames) != WizardStepCount {
		t.Errorf("wizardStepNames has %d items, want %d", len(wizardStepNames), WizardStepCount)
	}

	expectedNames := []string{"Welcome", "Paths", "Credentials", "Settings", "Summary"}
	for i, name := range expectedNames {
		if wizardStepNames[i] != name {
			t.Errorf("wizardStepNames[%d] = %q, want %q", i, wizardStepNames[i], name)
		}
	}
}
