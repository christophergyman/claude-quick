package tui

import (
	"fmt"
	"strings"

	"github.com/christophergyman/claude-quick/internal/constants"
	"github.com/christophergyman/claude-quick/internal/devcontainer"
)

// renderWithHeader creates a strings.Builder with the standard title header
// If subtitle is provided, it's also rendered below the title
func renderWithHeader(subtitle string) *strings.Builder {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("claude-quick"))
	if subtitle != "" {
		b.WriteString("\n")
		b.WriteString(SubtitleStyle.Render(subtitle))
	}
	b.WriteString("\n\n")
	return &b
}

// RenderContainerStarting renders the loading state while container starts
func RenderContainerStarting(projectName string, spinnerView string) string {
	return renderSpinnerWithHint(spinnerView, "Starting", projectName, "This may take a moment...")
}

// RenderError renders an error message
func RenderError(err error, hint string) string {
	b := renderWithHeader("")
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

// renderConfirmDialog renders a generic confirmation dialog
// entityType: "container", "tmux session", etc.
// labelType: "Project", "Session", etc.
func renderConfirmDialog(operation, entityType, labelType, name string) string {
	b := renderWithHeader("")
	actionText := "Stop"
	if operation == "restart" {
		actionText = "Restart"
	}
	b.WriteString(ErrorStyle.Render(fmt.Sprintf("%s %s?", actionText, entityType)))
	b.WriteString("\n\n")
	b.WriteString(labelType + ": ")
	b.WriteString(SuccessStyle.Render(name))
	b.WriteString("\n\n")
	b.WriteString(HelpStyle.Render("y: Confirm  n/Esc: Cancel"))
	return b.String()
}

// RenderConfirmDialog renders a confirmation dialog for stop/restart operations
func RenderConfirmDialog(operation, projectName string) string {
	return renderConfirmDialog(operation, "container", "Project", projectName)
}

// renderOperation renders a generic spinner operation view
// entityType is optional and appears before the name (e.g., "session" for tmux)
func renderOperation(operation, entityType, name, spinnerView string) string {
	b := renderWithHeader("")
	b.WriteString(SpinnerStyle.Render(spinnerView))
	if entityType != "" {
		b.WriteString(fmt.Sprintf(" %s %s ", operation, entityType))
	} else {
		b.WriteString(fmt.Sprintf(" %s ", operation))
	}
	b.WriteString(SuccessStyle.Render(name))
	b.WriteString("...")
	return b.String()
}

// RenderContainerOperation renders progress during stop/restart operations
func RenderContainerOperation(operation, projectName, spinnerView string) string {
	return renderOperation(operation, "", projectName, spinnerView)
}

// RenderDiscovering renders the project discovery loading state
func RenderDiscovering(spinnerView string) string {
	return renderSpinnerWithHint(spinnerView, "Discovering projects", "", "Searching for devcontainer.json files...")
}

// truncatePath shortens a path to fit within maxLen
func truncatePath(path string, maxLen int) string {
	if maxLen <= 0 {
		maxLen = constants.DefaultPathTruncateLen
	}
	if len(path) <= maxLen {
		return path
	}
	return "..." + path[len(path)-maxLen+3:]
}

// RenderRefreshingStatus renders the loading state while refreshing container status
func RenderRefreshingStatus(spinnerView string) string {
	return renderSpinnerAction(spinnerView, "Refreshing container status", "")
}

// RenderDashboard renders the container dashboard with status indicators
func RenderDashboard(instances []devcontainer.ContainerInstanceWithStatus, cursor int, width int) string {
	b := renderWithHeader("Container Dashboard")

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
		pathLine := "    " + DimmedStyle.Render(truncatePath(instance.Path, width-constants.PathTruncatePadding))
		b.WriteString(pathLine)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("↑/↓: Navigate  Enter: Connect  n: New Worktree  d: Delete Worktree"))
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("x: Stop  r: Restart  R: Refresh  ?: Config  q: Quit"))
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
	b := renderWithHeader("New Git Worktree")
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
	return renderSpinnerWithHint(spinnerView, "Creating worktree", branchName, "Running git worktree add...")
}

// RenderConfirmDeleteWorktree renders the confirmation dialog for deleting a worktree
func RenderConfirmDeleteWorktree(branchName string) string {
	b := renderWithHeader("")
	b.WriteString(ErrorStyle.Render("Delete worktree?"))
	b.WriteString("\n\n")
	b.WriteString("Branch: ")
	b.WriteString(SuccessStyle.Render(branchName))
	b.WriteString("\n\n")
	b.WriteString(DimmedStyle.Render("This will remove the worktree directory and branch"))
	b.WriteString("\n\n")
	b.WriteString(HelpStyle.Render("y: Confirm  n/Esc: Cancel"))
	return b.String()
}

// RenderDeletingWorktree renders the loading state while deleting a worktree
func RenderDeletingWorktree(branchName string, spinnerView string) string {
	return renderSpinnerWithHint(spinnerView, "Deleting worktree", branchName, "Running git worktree remove...")
}
