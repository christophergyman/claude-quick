package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/christophergyman/claude-quick/internal/tmux"
)

const newSessionOption = "[+ New Session]"

// RenderTmuxSelect renders the tmux session selection view
func RenderTmuxSelect(projectName string, sessions []tmux.Session, cursor int) string {
	var b strings.Builder

	title := TitleStyle.Render("claude-quick")
	subtitle := SubtitleStyle.Render("tmux Sessions in: " + projectName)

	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(subtitle)
	b.WriteString("\n\n")

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

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("↑/↓: Navigate  Enter: Select  x: Stop  r: Restart  ?: Config  q: Back"))
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("Tip: Detach from tmux with Ctrl+b d"))

	return b.String()
}

// RenderNewSessionInput renders the text input view for new session name
func RenderNewSessionInput(projectName string, ti textinput.Model) string {
	var b strings.Builder

	title := TitleStyle.Render("claude-quick")
	subtitle := SubtitleStyle.Render("New Session in: " + projectName)

	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(subtitle)
	b.WriteString("\n\n")

	b.WriteString("Enter session name:\n\n")
	b.WriteString(ti.View())
	b.WriteString("\n\n")
	b.WriteString(HelpStyle.Render("Enter: Create  Esc: Cancel"))
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("Tip: Detach from tmux with Ctrl+b d"))

	return b.String()
}

// RenderAttaching renders the view while attaching to a tmux session
func RenderAttaching(projectName, sessionName, spinnerView string) string {
	var b strings.Builder

	title := TitleStyle.Render("claude-quick")
	b.WriteString(title)
	b.WriteString("\n\n")

	b.WriteString(SpinnerStyle.Render(spinnerView))
	b.WriteString(" Attaching to ")
	b.WriteString(SuccessStyle.Render(sessionName))
	b.WriteString("...")

	return b.String()
}

// RenderLoadingTmuxSessions renders the loading state while fetching tmux sessions
func RenderLoadingTmuxSessions(projectName, spinnerView string) string {
	var b strings.Builder

	title := TitleStyle.Render("claude-quick")
	subtitle := SubtitleStyle.Render("Project: " + projectName)

	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(subtitle)
	b.WriteString("\n\n")

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
	var b strings.Builder

	title := TitleStyle.Render("claude-quick")
	b.WriteString(title)
	b.WriteString("\n\n")

	actionText := "Stop"
	if operation == "restart" {
		actionText = "Restart"
	}
	b.WriteString(ErrorStyle.Render(actionText + " tmux session?"))
	b.WriteString("\n\n")
	b.WriteString("Session: ")
	b.WriteString(SuccessStyle.Render(sessionName))
	b.WriteString("\n\n")
	b.WriteString(HelpStyle.Render("y: Confirm  n/Esc: Cancel"))

	return b.String()
}

// RenderTmuxOperation renders progress during tmux stop/restart operations
func RenderTmuxOperation(operation, sessionName, spinnerView string) string {
	var b strings.Builder

	title := TitleStyle.Render("claude-quick")
	b.WriteString(title)
	b.WriteString("\n\n")

	b.WriteString(SpinnerStyle.Render(spinnerView))
	b.WriteString(" " + operation + " session ")
	b.WriteString(SuccessStyle.Render(sessionName))
	b.WriteString("...")

	return b.String()
}
