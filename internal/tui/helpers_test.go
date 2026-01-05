package tui

import (
	"strings"
	"testing"

	"github.com/christophergyman/claude-quick/internal/devcontainer"
	"github.com/christophergyman/claude-quick/internal/tmux"
)

func TestTotalTmuxOptions(t *testing.T) {
	tests := []struct {
		name     string
		sessions []tmux.Session
		expected int
	}{
		{
			name:     "empty sessions",
			sessions: []tmux.Session{},
			expected: 1, // Just the "New Session" option
		},
		{
			name:     "one session",
			sessions: []tmux.Session{{Name: "main"}},
			expected: 2, // session + "New Session"
		},
		{
			name: "multiple sessions",
			sessions: []tmux.Session{
				{Name: "main"},
				{Name: "dev"},
				{Name: "test"},
			},
			expected: 4, // 3 sessions + "New Session"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TotalTmuxOptions(tt.sessions)
			if result != tt.expected {
				t.Errorf("TotalTmuxOptions() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestIsNewSessionSelected(t *testing.T) {
	tests := []struct {
		name     string
		sessions []tmux.Session
		cursor   int
		expected bool
	}{
		{
			name:     "cursor at new session with empty list",
			sessions: []tmux.Session{},
			cursor:   0,
			expected: true, // 0 == len([]) = 0, so new session is selected
		},
		{
			name:     "cursor at first session",
			sessions: []tmux.Session{{Name: "main"}},
			cursor:   0,
			expected: false, // 0 != len([main]) = 1
		},
		{
			name:     "cursor at new session with one session",
			sessions: []tmux.Session{{Name: "main"}},
			cursor:   1,
			expected: true, // 1 == len([main]) = 1
		},
		{
			name: "cursor in middle of list",
			sessions: []tmux.Session{
				{Name: "main"},
				{Name: "dev"},
				{Name: "test"},
			},
			cursor:   1,
			expected: false, // 1 != 3
		},
		{
			name: "cursor at new session with multiple sessions",
			sessions: []tmux.Session{
				{Name: "main"},
				{Name: "dev"},
				{Name: "test"},
			},
			cursor:   3,
			expected: true, // 3 == 3
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNewSessionSelected(tt.sessions, tt.cursor)
			if result != tt.expected {
				t.Errorf("IsNewSessionSelected() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTruncatePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		maxLen   int
		expected string
	}{
		{
			name:     "short path - no truncation",
			path:     "/home/user",
			maxLen:   20,
			expected: "/home/user",
		},
		{
			name:     "exact length - no truncation",
			path:     "/home/user/project",
			maxLen:   18,
			expected: "/home/user/project",
		},
		{
			name:     "long path - truncated",
			path:     "/home/user/projects/very/long/path/to/project",
			maxLen:   20,
			expected: "...g/path/to/project", // "..." (3) + last 17 chars = 20 total
		},
		{
			name:     "empty path",
			path:     "",
			maxLen:   20,
			expected: "",
		},
		{
			name:     "zero maxLen uses default",
			path:     "/short",
			maxLen:   0,
			expected: "/short", // Will use default (40) so won't truncate
		},
		{
			name:     "negative maxLen uses default",
			path:     "/short",
			maxLen:   -5,
			expected: "/short", // Will use default (40) so won't truncate
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncatePath(tt.path, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncatePath(%q, %d) = %q, want %q", tt.path, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestTruncatePath_PreservesEnd(t *testing.T) {
	path := "/home/user/projects/myapp/src/components/Button.tsx"
	maxLen := 30
	result := truncatePath(path, maxLen)

	// Should preserve the end of the path
	if !strings.HasSuffix(result, "Button.tsx") {
		t.Errorf("truncatePath should preserve end of path, got %q", result)
	}

	// Should start with ...
	if !strings.HasPrefix(result, "...") {
		t.Errorf("truncatePath should start with '...', got %q", result)
	}

	// Total length should be maxLen
	if len(result) != maxLen {
		t.Errorf("truncatePath length = %d, want %d", len(result), maxLen)
	}
}

func TestGetStatusText(t *testing.T) {
	tests := []struct {
		name     string
		status   devcontainer.ContainerStatus
		contains string
	}{
		{
			name:     "running status",
			status:   devcontainer.StatusRunning,
			contains: "running",
		},
		{
			name:     "stopped status",
			status:   devcontainer.StatusStopped,
			contains: "stopped",
		},
		{
			name:     "unknown status",
			status:   devcontainer.StatusUnknown,
			contains: "unknown",
		},
		{
			name:     "empty status treated as unknown",
			status:   devcontainer.ContainerStatus(""),
			contains: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStatusText(tt.status)
			// The result includes ANSI escape codes from lipgloss styling,
			// so we check for content containment
			if !strings.Contains(result, tt.contains) {
				t.Errorf("getStatusText(%q) = %q, want to contain %q", tt.status, result, tt.contains)
			}
		})
	}
}

func TestGetStatusIcon(t *testing.T) {
	tests := []struct {
		name   string
		status devcontainer.ContainerStatus
		icon   string
	}{
		{
			name:   "running status",
			status: devcontainer.StatusRunning,
			icon:   "●",
		},
		{
			name:   "stopped status",
			status: devcontainer.StatusStopped,
			icon:   "○",
		},
		{
			name:   "unknown status",
			status: devcontainer.StatusUnknown,
			icon:   "?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStatusIcon(tt.status)
			// The result includes ANSI escape codes, check for the icon character
			if !strings.Contains(result, tt.icon) {
				t.Errorf("getStatusIcon(%q) = %q, want to contain %q", tt.status, result, tt.icon)
			}
		})
	}
}

func TestRenderSpinnerAction(t *testing.T) {
	// Test the basic rendering function
	spinner := "⠋"
	action := "Starting"
	name := "myproject"

	result := renderSpinnerAction(spinner, action, name)

	// Should contain all components
	if !strings.Contains(result, spinner) {
		t.Error("renderSpinnerAction should contain spinner")
	}
	if !strings.Contains(result, action) {
		t.Error("renderSpinnerAction should contain action")
	}
	if !strings.Contains(result, name) {
		t.Error("renderSpinnerAction should contain name")
	}
	if !strings.HasSuffix(result, "...") {
		t.Error("renderSpinnerAction should end with ...")
	}
}

func TestRenderSpinnerAction_NoName(t *testing.T) {
	spinner := "⠋"
	action := "Refreshing"

	result := renderSpinnerAction(spinner, action, "")

	if !strings.Contains(result, action) {
		t.Error("renderSpinnerAction should contain action")
	}
	if !strings.HasSuffix(result, "...") {
		t.Error("renderSpinnerAction should end with ...")
	}
}

func TestRenderSpinnerWithHint(t *testing.T) {
	spinner := "⠋"
	action := "Starting"
	name := "myproject"
	hint := "This may take a while"

	result := renderSpinnerWithHint(spinner, action, name, hint)

	if !strings.Contains(result, hint) {
		t.Error("renderSpinnerWithHint should contain hint")
	}
	if !strings.Contains(result, action) {
		t.Error("renderSpinnerWithHint should contain action")
	}
}

func TestRenderSpinnerWithHint_NoHint(t *testing.T) {
	spinner := "⠋"
	action := "Starting"
	name := "myproject"

	result := renderSpinnerWithHint(spinner, action, name, "")

	// Should just be the spinner action without hint
	if !strings.Contains(result, action) {
		t.Error("renderSpinnerWithHint should contain action")
	}
}

// ============================================================================
// styles.go tests
// ============================================================================

func TestRepeatChar(t *testing.T) {
	tests := []struct {
		name     string
		char     string
		n        int
		expected string
	}{
		{"zero count", "x", 0, ""},
		{"negative count", "x", -5, ""},
		{"one char", "x", 1, "x"},
		{"multiple chars", "-", 5, "-----"},
		{"unicode char", "─", 3, "───"},
		{"multi-char string", "ab", 3, "ababab"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := repeatChar(tt.char, tt.n)
			if result != tt.expected {
				t.Errorf("repeatChar(%q, %d) = %q, want %q", tt.char, tt.n, result, tt.expected)
			}
		})
	}
}

func TestCursor(t *testing.T) {
	result := Cursor()

	// Should contain the cursor symbol
	if !strings.Contains(result, "›") {
		t.Errorf("Cursor() = %q, should contain '›'", result)
	}

	// Should have consistent length (styled, so may have ANSI codes)
	if len(result) == 0 {
		t.Error("Cursor() should not return empty string")
	}
}

func TestNoCursor(t *testing.T) {
	result := NoCursor()

	// Should be exactly 2 spaces
	if result != "  " {
		t.Errorf("NoCursor() = %q, want %q", result, "  ")
	}

	// Length should match Cursor's visual width (2 chars)
	if len(result) != 2 {
		t.Errorf("NoCursor() length = %d, want 2", len(result))
	}
}

func TestRenderSeparator(t *testing.T) {
	tests := []struct {
		name  string
		width int
	}{
		{"zero width uses default", 0},
		{"negative width uses default", -10},
		{"small width", 10},
		{"medium width", 40},
		{"large width", 80},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderSeparator(tt.width)

			// Should contain the separator character
			if !strings.Contains(result, "─") {
				t.Errorf("RenderSeparator(%d) should contain '─'", tt.width)
			}

			// Should not be empty
			if len(result) == 0 {
				t.Errorf("RenderSeparator(%d) should not be empty", tt.width)
			}
		})
	}
}

func TestRenderKeyBinding(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		description string
	}{
		{"simple key", "q", "quit"},
		{"arrow key", "↑↓", "navigate"},
		{"modifier key", "ctrl+c", "cancel"},
		{"enter key", "enter", "select"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderKeyBinding(tt.key, tt.description)

			// Should contain both key and description
			if !strings.Contains(result, tt.key) {
				t.Errorf("RenderKeyBinding(%q, %q) should contain key", tt.key, tt.description)
			}
			if !strings.Contains(result, tt.description) {
				t.Errorf("RenderKeyBinding(%q, %q) should contain description", tt.key, tt.description)
			}
		})
	}
}

func TestRenderBorderedHeader(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		subtitle string
		width    int
	}{
		{"basic header", "Title", "Subtitle", 60},
		{"no subtitle", "Title", "", 60},
		{"zero width uses default", "Title", "Sub", 0},
		{"narrow width", "Title", "Subtitle", 30},
		{"wide width", "Title", "Subtitle", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderBorderedHeader(tt.title, tt.subtitle, tt.width)

			// Should contain title
			if !strings.Contains(result, tt.title) {
				t.Errorf("RenderBorderedHeader should contain title %q", tt.title)
			}

			// Should contain subtitle if provided
			if tt.subtitle != "" && !strings.Contains(result, tt.subtitle) {
				t.Errorf("RenderBorderedHeader should contain subtitle %q", tt.subtitle)
			}

			// Should contain border characters
			if !strings.Contains(result, "┌") || !strings.Contains(result, "┐") {
				t.Error("RenderBorderedHeader should contain top border corners")
			}
			if !strings.Contains(result, "└") || !strings.Contains(result, "┘") {
				t.Error("RenderBorderedHeader should contain bottom border corners")
			}
			if !strings.Contains(result, "│") {
				t.Error("RenderBorderedHeader should contain side borders")
			}
		})
	}
}

func TestGetStatusIndicator(t *testing.T) {
	tests := []struct {
		name     string
		running  bool
		unknown  bool
		contains string
	}{
		{"running", true, false, "running"},
		{"stopped", false, false, "stopped"},
		{"unknown", false, true, "unknown"},
		{"unknown takes precedence", true, true, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetStatusIndicator(tt.running, tt.unknown)

			if !strings.Contains(result, tt.contains) {
				t.Errorf("GetStatusIndicator(%v, %v) = %q, should contain %q",
					tt.running, tt.unknown, result, tt.contains)
			}
		})
	}
}

// ============================================================================
// container.go tests
// ============================================================================

func TestRenderSimpleHeader(t *testing.T) {
	tests := []struct {
		name     string
		subtitle string
	}{
		{"with subtitle", "Container Dashboard"},
		{"empty subtitle", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderSimpleHeader(tt.subtitle)
			content := result.String()

			// Should contain the app title
			if !strings.Contains(content, "claude-quick") {
				t.Error("renderSimpleHeader should contain 'claude-quick'")
			}

			// Should contain subtitle if provided
			if tt.subtitle != "" && !strings.Contains(content, tt.subtitle) {
				t.Errorf("renderSimpleHeader should contain subtitle %q", tt.subtitle)
			}
		})
	}
}

func TestRenderWithHeader(t *testing.T) {
	tests := []struct {
		name     string
		subtitle string
	}{
		{"with subtitle", "Container Dashboard"},
		{"empty subtitle", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderWithHeader(tt.subtitle)
			content := result.String()

			// Should contain the app title
			if !strings.Contains(content, "claude-quick") {
				t.Error("renderWithHeader should contain 'claude-quick'")
			}

			// Should contain border characters (from RenderBorderedHeader)
			if !strings.Contains(content, "┌") {
				t.Error("renderWithHeader should contain bordered header")
			}
		})
	}
}

func TestRenderConfirmDialog(t *testing.T) {
	tests := []struct {
		name       string
		operation  string
		entityType string
		labelType  string
		entityName string
		contains   []string
	}{
		{
			name:       "stop container",
			operation:  "stop",
			entityType: "container",
			labelType:  "Project",
			entityName: "myproject",
			contains:   []string{"Stop", "container", "Project", "myproject", "Confirm", "Cancel"},
		},
		{
			name:       "restart container",
			operation:  "restart",
			entityType: "container",
			labelType:  "Project",
			entityName: "myapp",
			contains:   []string{"Restart", "container", "Project", "myapp"},
		},
		{
			name:       "stop tmux session",
			operation:  "stop",
			entityType: "tmux session",
			labelType:  "Session",
			entityName: "main",
			contains:   []string{"Stop", "tmux session", "Session", "main"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderConfirmDialog(tt.operation, tt.entityType, tt.labelType, tt.entityName)

			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("renderConfirmDialog should contain %q", expected)
				}
			}
		})
	}
}

func TestRenderOperation(t *testing.T) {
	tests := []struct {
		name       string
		operation  string
		entityType string
		entityName string
		spinner    string
	}{
		{"with entity type", "Stopping", "session", "main", "⠋"},
		{"without entity type", "Stopping", "", "myproject", "⠋"},
		{"restart operation", "Restarting", "", "myapp", "⠙"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderOperation(tt.operation, tt.entityType, tt.entityName, tt.spinner)

			// Should contain operation
			if !strings.Contains(result, tt.operation) {
				t.Errorf("renderOperation should contain operation %q", tt.operation)
			}

			// Should contain entity name
			if !strings.Contains(result, tt.entityName) {
				t.Errorf("renderOperation should contain entity name %q", tt.entityName)
			}

			// Should contain entity type if provided
			if tt.entityType != "" && !strings.Contains(result, tt.entityType) {
				t.Errorf("renderOperation should contain entity type %q", tt.entityType)
			}

			// Should contain spinner
			if !strings.Contains(result, tt.spinner) {
				t.Errorf("renderOperation should contain spinner %q", tt.spinner)
			}

			// Should end with ...
			if !strings.HasSuffix(result, "...") {
				t.Error("renderOperation should end with ...")
			}
		})
	}
}

// ============================================================================
// model.go tests
// ============================================================================

func TestModel_GetInstanceName(t *testing.T) {
	tests := []struct {
		name     string
		model    Model
		expected string
	}{
		{
			name:     "nil instance",
			model:    Model{selectedInstance: nil},
			expected: "",
		},
		{
			name: "valid instance",
			model: Model{
				selectedInstance: &devcontainer.ContainerInstance{
					Project: devcontainer.Project{Name: "myproject", Path: "/path/to/project"},
				},
			},
			expected: "myproject",
		},
		{
			name: "instance with worktree",
			model: Model{
				selectedInstance: &devcontainer.ContainerInstance{
					Project: devcontainer.Project{Name: "myproject", Path: "/path/to/project"},
					Worktree: &devcontainer.WorktreeInfo{
						Branch: "feature/auth",
						IsMain: false,
					},
				},
			},
			expected: "myproject [feature/auth]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.model.getInstanceName()
			if result != tt.expected {
				t.Errorf("getInstanceName() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestModel_GetSessionName(t *testing.T) {
	tests := []struct {
		name     string
		model    Model
		expected string
	}{
		{
			name:     "nil session",
			model:    Model{selectedSession: nil},
			expected: "",
		},
		{
			name: "valid session",
			model: Model{
				selectedSession: &tmux.Session{Name: "main"},
			},
			expected: "main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.model.getSessionName()
			if result != tt.expected {
				t.Errorf("getSessionName() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestModel_GetWorktreeBranch(t *testing.T) {
	tests := []struct {
		name     string
		model    Model
		expected string
	}{
		{
			name:     "nil instance",
			model:    Model{selectedInstance: nil},
			expected: "",
		},
		{
			name: "instance without worktree",
			model: Model{
				selectedInstance: &devcontainer.ContainerInstance{
					Project:  devcontainer.Project{Name: "myproject"},
					Worktree: nil,
				},
			},
			expected: "",
		},
		{
			name: "instance with worktree",
			model: Model{
				selectedInstance: &devcontainer.ContainerInstance{
					Project: devcontainer.Project{Name: "myproject"},
					Worktree: &devcontainer.WorktreeInfo{
						Branch: "feature/auth",
					},
				},
			},
			expected: "feature/auth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.model.getWorktreeBranch()
			if result != tt.expected {
				t.Errorf("getWorktreeBranch() = %q, want %q", result, tt.expected)
			}
		})
	}
}
