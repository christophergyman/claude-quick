package tui

import (
	"fmt"
	"strings"

	"github.com/chezu/quickvibe/internal/devcontainer"
)

// RenderContainerSelect renders the container selection view
func RenderContainerSelect(projects []devcontainer.Project, cursor int, width int) string {
	var b strings.Builder

	title := TitleStyle.Render("quickVibe")
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
		b.WriteString(DimmedStyle.Render("~/.config/quickvibe/config.yaml"))
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
	b.WriteString(HelpStyle.Render("↑/↓: Navigate  Enter: Select  q: Quit"))

	return b.String()
}

// RenderContainerStarting renders the loading state while container starts
func RenderContainerStarting(projectName string, spinnerView string) string {
	var b strings.Builder

	title := TitleStyle.Render("quickVibe")
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

	title := TitleStyle.Render("quickVibe")
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
