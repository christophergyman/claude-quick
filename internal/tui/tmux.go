package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/christophergyman/claude-quick/internal/tmux"
)

const newSessionOption = "[+ New Session]"

// RenderTmuxSelect renders the tmux session selection view
func RenderTmuxSelect(projectName string, sessions []tmux.Session, cursor int, authWarning string) string {
	width := defaultWidth

	var b strings.Builder

	// Bordered header
	b.WriteString(RenderBorderedHeader("claude-quick", "tmux Sessions: "+projectName, width))
	b.WriteString("\n\n")

	// Show auth warning if present
	if authWarning != "" {
		b.WriteString(WarningStyle.Render("Auth warning: " + authWarning))
		b.WriteString("\n\n")
	}

	// Column headers
	sessionHeader := ColumnHeaderStyle.Render("SESSIONS")
	b.WriteString("  " + sessionHeader)
	b.WriteString("\n")

	// Separator line
	b.WriteString("  " + RenderSeparator(width-4))
	b.WriteString("\n")

	// Render sessions
	for i, session := range sessions {
		var line string
		display := session.FormatSession()
		if i == cursor {
			line = Cursor() + SelectedStyle.Render(display)
		} else {
			line = NoCursor() + ItemStyle.Render(display)
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	// Render "New Session" option
	newSessionIdx := len(sessions)
	var newSessionLine string
	if cursor == newSessionIdx {
		newSessionLine = Cursor() + SelectedStyle.Render(newSessionOption)
	} else {
		newSessionLine = NoCursor() + ItemStyle.Render(newSessionOption)
	}
	b.WriteString(newSessionLine)
	b.WriteString("\n")

	// Footer section
	b.WriteString("\n")
	b.WriteString("  " + RenderSeparator(width-4))
	b.WriteString("\n")

	// Key bindings - first row
	keybindings1 := fmt.Sprintf("  %s  %s  %s  %s",
		RenderKeyBinding("↑↓", "navigate"),
		RenderKeyBinding("enter", "select"),
		RenderKeyBinding("x", "stop"),
		RenderKeyBinding("r", "restart"),
	)
	b.WriteString(keybindings1)
	b.WriteString("\n")

	// Key bindings - second row with right-aligned detach hint
	leftKeys := fmt.Sprintf("  %s  %s  %s",
		RenderKeyBinding("t", "theme"),
		RenderKeyBinding("?", "config"),
		RenderKeyBinding("q", "back"),
	)
	rightKey := RenderKeyBinding("ctrl+b d", "detach")
	// Calculate spacing for right alignment
	leftWidth := lipgloss.Width(leftKeys)
	rightWidth := lipgloss.Width(rightKey)
	footerSpacing := width - leftWidth - rightWidth - 2
	if footerSpacing < 1 {
		footerSpacing = 1
	}
	b.WriteString(leftKeys + repeatChar(" ", footerSpacing) + rightKey)

	return b.String()
}

// RenderNewSessionInput renders the text input view for new session name
func RenderNewSessionInput(projectName string, ti textinput.Model) string {
	b := renderWithHeader("New Session: " + projectName)
	b.WriteString("Enter session name:")
	b.WriteString("\n\n")
	b.WriteString(ti.View())
	b.WriteString("\n\n")

	// Footer
	b.WriteString(RenderSeparator(defaultWidth - 4))
	b.WriteString("\n")
	keybindings := fmt.Sprintf("%s  %s",
		RenderKeyBinding("enter", "create"),
		RenderKeyBinding("esc", "cancel"),
	)
	b.WriteString(keybindings)

	return b.String()
}

// RenderAttaching renders the view while attaching to a tmux session
func RenderAttaching(projectName, sessionName, spinnerView string) string {
	return renderSpinnerAction(spinnerView, "Attaching to", sessionName)
}

// RenderLoadingTmuxSessions renders the loading state while fetching tmux sessions
func RenderLoadingTmuxSessions(projectName, spinnerView string) string {
	b := renderSimpleHeader("Project: " + projectName)
	b.WriteString(SpinnerStyle.Render(spinnerView))
	b.WriteString(" Loading tmux sessions...")
	return b.String()
}

// TotalTmuxOptions returns the total number of selectable options (sessions + new session)
func TotalTmuxOptions(sessions []tmux.Session) int {
	return len(sessions) + 1
}

// IsNewSessionSelected returns true if the cursor is on the "New Session" option
func IsNewSessionSelected(sessions []tmux.Session, cursor int) bool {
	return cursor == len(sessions)
}

// RenderTmuxConfirmDialog renders a confirmation dialog for tmux stop/restart operations
func RenderTmuxConfirmDialog(operation, sessionName string) string {
	return renderConfirmDialog(operation, "tmux session", "Session", sessionName)
}

// RenderTmuxOperation renders progress during tmux stop/restart operations
func RenderTmuxOperation(operation, sessionName, spinnerView string) string {
	return renderOperation(operation, "session", sessionName, spinnerView)
}
