package tui

import (
	"fmt"
	"strings"

	"github.com/christophergyman/claude-quick/internal/devcontainer"
)

// RenderContainerSelect renders the container selection view
func RenderContainerSelect(projects []devcontainer.Project, cursor int, width int) string {
	var b strings.Builder

	title := TitleStyle.Render("claude-quick")
	subtitle := SubtitleStyle.Render("Select Dev Container")

	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(subtitle)
	b.WriteString("\n\n")

	if len(projects) == 0 {
		b.WriteString(ErrorStyle.Render("No devcontainer projects found."))
		b.WriteString("\n\n")
		b.WriteString(DimmedStyle.Render("Add search paths to: "))
		b.WriteString("\n")
		b.WriteString(DimmedStyle.Render("~/.config/claude-quick/config.yaml"))
		return b.String()
	}

	for i, project := range projects {
		var line string
		if i == cursor {
			line = Cursor() + SelectedStyle.Render(project.Name)
		} else {
			line = NoCursor() + ItemStyle.Render(project.Name)
		}
		b.WriteString(line)
		b.WriteString("\n")

		// Show path on next line (dimmed)
		pathLine := "    " + DimmedStyle.Render(truncatePath(project.Path, width-6))
		b.WriteString(pathLine)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("↑/↓: Navigate  Enter: Select  x: Stop  r: Restart  ?: Config  q: Quit"))
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("Tip: Detach from tmux with Ctrl+b d"))

	return b.String()
}

// RenderContainerStarting renders the loading state while container starts
func RenderContainerStarting(projectName string, spinnerView string) string {
	var b strings.Builder

	title := TitleStyle.Render("claude-quick")
	b.WriteString(title)
	b.WriteString("\n\n")

	b.WriteString(SpinnerStyle.Render(spinnerView))
	b.WriteString(" Starting ")
	b.WriteString(SuccessStyle.Render(projectName))
	b.WriteString("...")
	b.WriteString("\n\n")
	b.WriteString(DimmedStyle.Render("This may take a moment..."))

	return b.String()
}

// RenderError renders an error message
func RenderError(err error, hint string) string {
	var b strings.Builder

	title := TitleStyle.Render("claude-quick")
	b.WriteString(title)
	b.WriteString("\n\n")

	b.WriteString(ErrorStyle.Render("Error: "))
	b.WriteString(fmt.Sprintf("%v", err))
	b.WriteString("\n\n")

	if hint != "" {
		b.WriteString(DimmedStyle.Render(hint))
		b.WriteString("\n\n")
	}

	b.WriteString(HelpStyle.Render("Press any key to continue"))

	return b.String()
}

// RenderConfirmDialog renders a confirmation dialog for stop/restart operations
func RenderConfirmDialog(operation, projectName string) string {
	var b strings.Builder

	title := TitleStyle.Render("claude-quick")
	b.WriteString(title)
	b.WriteString("\n\n")

	actionText := "Stop"
	if operation == "restart" {
		actionText = "Restart"
	}
	b.WriteString(ErrorStyle.Render(fmt.Sprintf("%s container?", actionText)))
	b.WriteString("\n\n")
	b.WriteString("Project: ")
	b.WriteString(SuccessStyle.Render(projectName))
	b.WriteString("\n\n")
	b.WriteString(HelpStyle.Render("y: Confirm  n/Esc: Cancel"))

	return b.String()
}

// RenderContainerOperation renders progress during stop/restart operations
func RenderContainerOperation(operation, projectName, spinnerView string) string {
	var b strings.Builder

	title := TitleStyle.Render("claude-quick")
	b.WriteString(title)
	b.WriteString("\n\n")

	b.WriteString(SpinnerStyle.Render(spinnerView))
	b.WriteString(fmt.Sprintf(" %s ", operation))
	b.WriteString(SuccessStyle.Render(projectName))
	b.WriteString("...")

	return b.String()
}

// RenderDiscovering renders the project discovery loading state
func RenderDiscovering(spinnerView string) string {
	var b strings.Builder

	title := TitleStyle.Render("claude-quick")
	b.WriteString(title)
	b.WriteString("\n\n")

	b.WriteString(SpinnerStyle.Render(spinnerView))
	b.WriteString(" Discovering projects...")
	b.WriteString("\n\n")
	b.WriteString(DimmedStyle.Render("Searching for devcontainer.json files..."))

	return b.String()
}

// truncatePath shortens a path to fit within maxLen
func truncatePath(path string, maxLen int) string {
	if maxLen <= 0 {
		maxLen = 40
	}
	if len(path) <= maxLen {
		return path
	}
	return "..." + path[len(path)-maxLen+3:]
}
