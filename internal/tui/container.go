package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/christophergyman/claude-quick/internal/constants"
	"github.com/christophergyman/claude-quick/internal/devcontainer"
)

const defaultWidth = 65

// renderWithHeader creates a strings.Builder with the bordered header
func renderWithHeader(subtitle string) *strings.Builder {
	var b strings.Builder
	b.WriteString(RenderBorderedHeader("claude-quick", subtitle, defaultWidth))
	b.WriteString("\n\n")
	return &b
}

// renderSimpleHeader creates a header without the bordered box (for spinner views)
func renderSimpleHeader(subtitle string) *strings.Builder {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("claude-quick"))
	b.WriteString("\n")
	if subtitle != "" {
		b.WriteString(SubtitleStyle.Render(subtitle))
		b.WriteString("\n")
	}
	b.WriteString("\n")
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
	b := renderSimpleHeader("")
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
	if width <= 0 {
		width = defaultWidth
	}

	var b strings.Builder

	// Bordered header
	b.WriteString(RenderBorderedHeader("claude-quick", "Container Dashboard", width))
	b.WriteString("\n\n")

	if len(instances) == 0 {
		b.WriteString(ErrorStyle.Render("No devcontainer projects found."))
		b.WriteString("\n\n")
		b.WriteString(DimmedStyle.Render("Add search paths to: "))
		b.WriteString("\n")
		b.WriteString(DimmedStyle.Render("~/.config/claude-quick/config.yaml"))
		return b.String()
	}

	// Column headers
	projectHeader := ColumnHeaderStyle.Render("PROJECTS")
	statusHeader := ColumnHeaderStyle.Render("STATUS")
	// Calculate spacing for right-aligned STATUS header
	statusWidth := lipgloss.Width(statusHeader)
	headerSpacing := width - 4 - lipgloss.Width(projectHeader) - statusWidth
	if headerSpacing < 1 {
		headerSpacing = 1
	}
	b.WriteString("  " + projectHeader + repeatChar(" ", headerSpacing) + statusHeader)
	b.WriteString("\n")

	// Separator line
	b.WriteString("  " + RenderSeparator(width-4))
	b.WriteString("\n")

	// Render each project
	for i, instance := range instances {
		// Get status indicator with text
		statusText := getStatusText(instance.Status)
		statusWidth := lipgloss.Width(statusText)

		// Session info for running containers
		sessionInfo := ""
		if instance.Status == devcontainer.StatusRunning && instance.SessionCount > 0 {
			sessionInfo = fmt.Sprintf(" [%d]", instance.SessionCount)
		}

		// Project name
		displayName := instance.DisplayName() + sessionInfo

		// Calculate spacing for right alignment
		nameWidth := lipgloss.Width(displayName)
		spacing := width - 4 - nameWidth - statusWidth
		if spacing < 1 {
			spacing = 1
		}

		// Render project line
		var line string
		if i == cursor {
			line = Cursor() + SelectedStyle.Render(displayName) + repeatChar(" ", spacing) + statusText
		} else {
			line = NoCursor() + ItemStyle.Render(displayName) + repeatChar(" ", spacing) + statusText
		}
		b.WriteString(line)
		b.WriteString("\n")

		// Show path on next line (dimmed, indented)
		pathLine := "    " + DimmedStyle.Render(truncatePath(instance.Path, width-constants.PathTruncatePadding))
		b.WriteString(pathLine)
		b.WriteString("\n")

		// Add spacing between entries except for the last one
		if i < len(instances)-1 {
			b.WriteString("\n")
		}
	}

	// Footer section
	b.WriteString("\n")
	b.WriteString("  " + RenderSeparator(width-4))
	b.WriteString("\n")

	// Key bindings - first row
	keybindings1 := fmt.Sprintf("  %s  %s  %s  %s  %s  %s",
		RenderKeyBinding("↑↓", "navigate"),
		RenderKeyBinding("enter", "connect"),
		RenderKeyBinding("n", "new"),
		RenderKeyBinding("d", "delete"),
		RenderKeyBinding("x", "stop"),
		RenderKeyBinding("r", "restart"),
	)
	b.WriteString(keybindings1)
	b.WriteString("\n")

	// Key bindings - second row with right-aligned detach hint
	leftKeys := fmt.Sprintf("  %s  %s  %s",
		RenderKeyBinding("R", "refresh"),
		RenderKeyBinding("?", "config"),
		RenderKeyBinding("q", "quit"),
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

// getStatusText returns a visual indicator with text label for container status
func getStatusText(status devcontainer.ContainerStatus) string {
	switch status {
	case devcontainer.StatusRunning:
		return StatusRunning.Render("● running")
	case devcontainer.StatusStopped:
		return StatusStopped.Render("○ stopped")
	default:
		return StatusUnknown.Render("? unknown")
	}
}

// getStatusIcon returns just the icon (for backward compatibility)
func getStatusIcon(status devcontainer.ContainerStatus) string {
	switch status {
	case devcontainer.StatusRunning:
		return SuccessStyle.Render("●")
	case devcontainer.StatusStopped:
		return WarningStyle.Render("○")
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
