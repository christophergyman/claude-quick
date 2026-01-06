package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/christophergyman/claude-quick/internal/github"
)

// RenderGitHubIssuesLoading renders the loading state while fetching issues
func RenderGitHubIssuesLoading(spinnerView string) string {
	return renderSpinnerWithHint(spinnerView, "Fetching GitHub issues", "", "Querying repository issues via gh CLI...")
}

// RenderGitHubIssuesList renders the GitHub issues list view
func RenderGitHubIssuesList(issues []github.Issue, cursor int, repoOwner, repoName string, width int) string {
	if width <= 0 {
		width = defaultWidth
	}

	var b strings.Builder

	// Header
	subtitle := fmt.Sprintf("GitHub Issues: %s/%s", repoOwner, repoName)
	b.WriteString(RenderBorderedHeader("claude-quick", subtitle, width))
	b.WriteString("\n\n")

	if len(issues) == 0 {
		b.WriteString(DimmedStyle.Render("No issues found."))
		b.WriteString("\n\n")
		b.WriteString(DimmedStyle.Render("Try changing the default_state filter in config."))
		b.WriteString("\n")
	} else {
		// Column headers
		b.WriteString("  ")
		b.WriteString(ColumnHeaderStyle.Render("#"))
		b.WriteString("      ")
		b.WriteString(ColumnHeaderStyle.Render("TITLE"))
		stateHeader := ColumnHeaderStyle.Render("STATE")
		// Right-align STATE header
		headerLen := lipgloss.Width("#      TITLE")
		spacing := width - 4 - headerLen - lipgloss.Width(stateHeader)
		if spacing < 1 {
			spacing = 1
		}
		b.WriteString(repeatChar(" ", spacing))
		b.WriteString(stateHeader)
		b.WriteString("\n")

		// Separator
		b.WriteString("  " + RenderSeparator(width-4))
		b.WriteString("\n")

		// Render issues
		for i, issue := range issues {
			renderIssueRow(&b, issue, i == cursor, width)
		}
	}

	// Footer
	b.WriteString("\n")
	b.WriteString("  " + RenderSeparator(width-4))
	b.WriteString("\n")

	// Key bindings
	keybindings := fmt.Sprintf("  %s  %s  %s  %s  %s",
		RenderKeyBinding("↑↓", "navigate"),
		RenderKeyBinding("enter", "create worktree"),
		RenderKeyBinding("v", "view"),
		RenderKeyBinding("r", "refresh"),
		RenderKeyBinding("q", "back"),
	)
	b.WriteString(keybindings)

	return b.String()
}

// renderIssueRow renders a single issue row
func renderIssueRow(b *strings.Builder, issue github.Issue, selected bool, width int) {
	// Format: #123   Title truncated...            open
	numberStr := fmt.Sprintf("#%-5d", issue.Number)

	// State indicator with priority: in-progress > open > closed
	var stateIndicator string
	if issue.HasLabel("in-progress") {
		stateIndicator = StatusInProgress.Render("in-progress")
	} else if issue.State == github.IssueStateOpen {
		stateIndicator = StatusRunning.Render("open")
	} else {
		stateIndicator = StatusStopped.Render("closed")
	}
	stateWidth := lipgloss.Width(stateIndicator)

	// Calculate title width
	numberWidth := lipgloss.Width(numberStr)
	titleMaxWidth := width - 4 - numberWidth - 2 - stateWidth - 2
	if titleMaxWidth < 10 {
		titleMaxWidth = 10 // Minimum reasonable width for title
	}

	title := issue.Title
	if len(title) > titleMaxWidth {
		title = title[:titleMaxWidth-3] + "..."
	}

	// Calculate spacing for right-aligned state
	titleWidth := lipgloss.Width(title)
	spacing := width - 4 - numberWidth - 2 - titleWidth - stateWidth
	if spacing < 1 {
		spacing = 1
	}

	// Build line
	if selected {
		b.WriteString(Cursor())
		b.WriteString(SelectedStyle.Render(numberStr))
		b.WriteString("  ")
		b.WriteString(SelectedStyle.Render(title))
	} else {
		b.WriteString(NoCursor())
		b.WriteString(ItemStyle.Render(numberStr))
		b.WriteString("  ")
		b.WriteString(ItemStyle.Render(title))
	}
	b.WriteString(repeatChar(" ", spacing))
	b.WriteString(stateIndicator)
	b.WriteString("\n")
}

// RenderGitHubIssueDetail renders the detailed view of a single issue
func RenderGitHubIssueDetail(issue *github.Issue, body string, width int) string {
	if width <= 0 {
		width = defaultWidth
	}

	// Handle nil issue
	if issue == nil {
		return RenderError(fmt.Errorf("no issue to display"), "Press any key to go back")
	}

	var b strings.Builder

	// Header
	subtitle := fmt.Sprintf("Issue #%d", issue.Number)
	b.WriteString(RenderBorderedHeader("claude-quick", subtitle, width))
	b.WriteString("\n\n")

	// Issue title
	b.WriteString(SelectedStyle.Render(issue.Title))
	b.WriteString("\n\n")

	// State indicator with priority: in-progress > open > closed
	var stateText string
	if issue.HasLabel("in-progress") {
		stateText = StatusInProgress.Render("◆ in-progress")
	} else if issue.State == github.IssueStateOpen {
		stateText = StatusRunning.Render("● open")
	} else {
		stateText = StatusStopped.Render("○ closed")
	}
	b.WriteString("State: ")
	b.WriteString(stateText)
	b.WriteString("\n\n")

	// Separator
	b.WriteString(RenderSeparator(width - 4))
	b.WriteString("\n\n")

	// Issue body (wrapped and truncated if needed)
	if body == "" {
		b.WriteString(DimmedStyle.Render("No description provided."))
	} else {
		// Simple word wrap for body text
		wrappedBody := wrapText(body, width-4)
		// Limit to reasonable height
		lines := strings.Split(wrappedBody, "\n")
		maxLines := 15
		if len(lines) > maxLines {
			lines = lines[:maxLines]
			lines = append(lines, DimmedStyle.Render("... (truncated)"))
		}
		b.WriteString(strings.Join(lines, "\n"))
	}

	b.WriteString("\n\n")

	// Footer
	b.WriteString(RenderSeparator(width - 4))
	b.WriteString("\n")

	// Key bindings
	keybindings := fmt.Sprintf("  %s  %s  %s",
		RenderKeyBinding("enter", "create worktree"),
		RenderKeyBinding("t", "theme"),
		RenderKeyBinding("q", "back"),
	)
	b.WriteString(keybindings)

	return b.String()
}

// RenderGitHubWorktreeCreating renders the loading state during worktree creation from issue
func RenderGitHubWorktreeCreating(issueNumber int, spinnerView string) string {
	return renderSpinnerWithHint(spinnerView,
		fmt.Sprintf("Creating worktree for issue #%d", issueNumber),
		"",
		"Running git worktree add...")
}

// RenderGitHubIssueDetailLoading renders the loading state while fetching issue details
func RenderGitHubIssueDetailLoading(issueNumber int, spinnerView string) string {
	return renderSpinnerWithHint(spinnerView,
		fmt.Sprintf("Loading issue #%d", issueNumber),
		"",
		"Fetching issue details...")
}

// wrapText wraps text at word boundaries to fit within maxWidth
func wrapText(text string, maxWidth int) string {
	if maxWidth <= 0 {
		maxWidth = 60
	}

	var result strings.Builder
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}

		// Handle empty lines
		if len(line) == 0 {
			continue
		}

		// Simple word wrap
		words := strings.Fields(line)
		currentLine := ""
		for _, word := range words {
			if currentLine == "" {
				currentLine = word
			} else if len(currentLine)+1+len(word) <= maxWidth {
				currentLine += " " + word
			} else {
				result.WriteString(currentLine)
				result.WriteString("\n")
				currentLine = word
			}
		}
		if currentLine != "" {
			result.WriteString(currentLine)
		}
	}

	return result.String()
}
