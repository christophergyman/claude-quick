package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/chezu/quickvibe/internal/tmux"
)

const newSessionOption = "[+ New Session]"

// RenderTmuxSelect renders the tmux session selection view
func RenderTmuxSelect(projectName string, sessions []tmux.Session, cursor int) string {
	var b strings.Builder

	title := TitleStyle.Render("quickVibe")
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
	b.WriteString(HelpStyle.Render("↑/↓: Navigate  Enter: Select  q: Back"))

	return b.String()
}

// RenderNewSessionInput renders the text input view for new session name
func RenderNewSessionInput(projectName string, ti textinput.Model) string {
	var b strings.Builder

	title := TitleStyle.Render("quickVibe")
	subtitle := SubtitleStyle.Render("New Session in: " + projectName)

	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(subtitle)
	b.WriteString("\n\n")

	b.WriteString("Enter session name:\n\n")
	b.WriteString(ti.View())
	b.WriteString("\n\n")
	b.WriteString(HelpStyle.Render("Enter: Create  Esc: Cancel"))

	return b.String()
}

// RenderAttaching renders the view while attaching to a tmux session
func RenderAttaching(projectName, sessionName, spinnerView string) string {
	var b strings.Builder

	title := TitleStyle.Render("quickVibe")
	b.WriteString(title)
	b.WriteString("\n\n")

	b.WriteString(SpinnerStyle.Render(spinnerView))
	b.WriteString(" Attaching to ")
	b.WriteString(SuccessStyle.Render(sessionName))
	b.WriteString("...")

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
