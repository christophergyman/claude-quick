package tui

import (
	"fmt"
	"strings"

	"github.com/christophergyman/claude-quick/internal/devcontainer"
)

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

// RenderRefreshingStatus renders the loading state while refreshing container status
func RenderRefreshingStatus(spinnerView string) string {
	var b strings.Builder

	title := TitleStyle.Render("claude-quick")
	b.WriteString(title)
	b.WriteString("\n\n")

	b.WriteString(SpinnerStyle.Render(spinnerView))
	b.WriteString(" Refreshing container status...")

	return b.String()
}

// RenderDashboard renders the container dashboard with status indicators
func RenderDashboard(instances []devcontainer.ContainerInstanceWithStatus, cursor int, width int) string {
	var b strings.Builder

	title := TitleStyle.Render("claude-quick")
	subtitle := SubtitleStyle.Render("Container Dashboard")

	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(subtitle)
	b.WriteString("\n\n")

	if len(instances) == 0 {
		b.WriteString(ErrorStyle.Render("No devcontainer projects found."))
		b.WriteString("\n\n")
		b.WriteString(DimmedStyle.Render("Add search paths to: "))
		b.WriteString("\n")
		b.WriteString(DimmedStyle.Render("~/.config/claude-quick/config.yaml"))
		return b.String()
	}

	for i, instance := range instances {
		// Status indicator
		statusIcon := getStatusIcon(instance.Status)

		// Session info for running containers
		sessionInfo := ""
		if instance.Status == devcontainer.StatusRunning && instance.SessionCount > 0 {
			sessionInfo = fmt.Sprintf(" [%d sessions]", instance.SessionCount)
		}

		// Use DisplayName() to show instance name (e.g., "project (claude1)")
		display := fmt.Sprintf("%s %s%s", statusIcon, instance.DisplayName(), sessionInfo)

		var line string
		if i == cursor {
			line = Cursor() + SelectedStyle.Render(display)
		} else {
			line = NoCursor() + ItemStyle.Render(display)
		}
		b.WriteString(line)
		b.WriteString("\n")

		// Show path on next line (dimmed)
		pathLine := "    " + DimmedStyle.Render(truncatePath(instance.Path, width-6))
		b.WriteString(pathLine)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("↑/↓: Navigate  Enter: Connect  n: New Worktree  x: Stop  r: Restart  R: Refresh  ?: Config  q: Quit"))
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("Tip: Detach from tmux with Ctrl+b d to return here"))

	return b.String()
}

// getStatusIcon returns a visual indicator for container status
func getStatusIcon(status devcontainer.ContainerStatus) string {
	switch status {
	case devcontainer.StatusRunning:
		return SuccessStyle.Render("●")
	case devcontainer.StatusStopped:
		return ErrorStyle.Render("○")
	default:
		return DimmedStyle.Render("?")
	}
}

// RenderNewWorktreeInput renders the text input for creating a new worktree
func RenderNewWorktreeInput(projectName string, input interface{ View() string }) string {
	var b strings.Builder

	title := TitleStyle.Render("claude-quick")
	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(SubtitleStyle.Render("New Git Worktree"))
	b.WriteString("\n\n")

	b.WriteString("Project: ")
	b.WriteString(SuccessStyle.Render(projectName))
	b.WriteString("\n\n")

	b.WriteString("Enter branch name (e.g., feature-auth, bugfix-123):")
	b.WriteString("\n\n")

	b.WriteString(input.View())
	b.WriteString("\n\n")

	b.WriteString(DimmedStyle.Render("Will create worktree in sibling directory with new branch"))
	b.WriteString("\n\n")
	b.WriteString(HelpStyle.Render("Enter: Create  Esc: Cancel"))

	return b.String()
}

// RenderCreatingWorktree renders the loading state while creating a new worktree
func RenderCreatingWorktree(branchName string, spinnerView string) string {
	var b strings.Builder

	title := TitleStyle.Render("claude-quick")
	b.WriteString(title)
	b.WriteString("\n\n")

	b.WriteString(SpinnerStyle.Render(spinnerView))
	b.WriteString(" Creating worktree ")
	b.WriteString(SuccessStyle.Render(branchName))
	b.WriteString("...")
	b.WriteString("\n\n")
	b.WriteString(DimmedStyle.Render("Running git worktree add..."))

	return b.String()
}
