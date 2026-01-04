package tui

import "github.com/charmbracelet/lipgloss"

// Colors - Claude Code inspired palette
var (
	colorOrange    = lipgloss.Color("#E07A5F") // Anthropic orange (terracotta/salmon)
	colorWhite     = lipgloss.Color("#FFFFFF") // White for selected items
	colorDim       = lipgloss.Color("#6B7280") // Gray for dimmed text
	colorSuccess   = lipgloss.Color("#10B981") // Green for running
	colorWarning   = lipgloss.Color("#F59E0B") // Yellow/amber for stopped
	colorError     = lipgloss.Color("#EF4444") // Red for errors
	colorSeparator = lipgloss.Color("#4B5563") // Darker gray for separators
)

// Styles
var (
	// Header title style (Anthropic orange)
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorOrange)

	// Subtitle style
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	// Selected item style - white bold (no background)
	SelectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite)

	// Unselected item style
	ItemStyle = lipgloss.NewStyle().
			Foreground(colorWhite)

	// Dimmed item style (for paths, details)
	DimmedStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	// Help text style
	HelpStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	// Error style
	ErrorStyle = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true)

	// Success style
	SuccessStyle = lipgloss.NewStyle().
			Foreground(colorSuccess)

	// Warning style
	WarningStyle = lipgloss.NewStyle().
			Foreground(colorWarning)

	// Box style for containers
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorDim).
			Padding(0, 1)

	// Input style for text input
	InputStyle = lipgloss.NewStyle().
			Foreground(colorWhite)

	// Spinner style
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(colorOrange)

	// Separator style
	SeparatorStyle = lipgloss.NewStyle().
			Foreground(colorSeparator)

	// Key highlight style (for keybinding display)
	KeyStyle = lipgloss.NewStyle().
			Foreground(colorWhite)

	// Column header style
	ColumnHeaderStyle = lipgloss.NewStyle().
				Foreground(colorDim).
				Bold(true)
)

// Status text styles with labels
var (
	StatusRunning = lipgloss.NewStyle().Foreground(colorSuccess)
	StatusStopped = lipgloss.NewStyle().Foreground(colorWarning)
	StatusUnknown = lipgloss.NewStyle().Foreground(colorDim)
)

// Cursor returns the selection cursor (› instead of >)
func Cursor() string {
	return lipgloss.NewStyle().
		Foreground(colorOrange).
		Bold(true).
		Render("› ")
}

// NoCursor returns spacing for non-selected items
func NoCursor() string {
	return "  "
}

// RenderSeparator returns a horizontal separator line
func RenderSeparator(width int) string {
	if width <= 0 {
		width = 60
	}
	line := ""
	for i := 0; i < width; i++ {
		line += "─"
	}
	return SeparatorStyle.Render(line)
}

// RenderKeyBinding formats a key binding with highlighted key
func RenderKeyBinding(key, description string) string {
	return KeyStyle.Render(key) + " " + DimmedStyle.Render(description)
}

// RenderBorderedHeader creates a bordered header box
func RenderBorderedHeader(title, subtitle string, width int) string {
	if width <= 0 {
		width = 60
	}
	// Inner width accounting for border and padding
	innerWidth := width - 4

	// Top border
	top := "┌" + repeatChar("─", width-2) + "┐"

	// Title line (left-padded)
	titleContent := "  " + TitleStyle.Render(title)
	titlePadding := innerWidth - lipgloss.Width(titleContent) + 2
	if titlePadding < 0 {
		titlePadding = 0
	}

	// Subtitle line
	subtitleContent := "  " + SubtitleStyle.Render(subtitle)
	subtitlePadding := innerWidth - lipgloss.Width(subtitleContent) + 2
	if subtitlePadding < 0 {
		subtitlePadding = 0
	}

	// Bottom border
	bottom := "└" + repeatChar("─", width-2) + "┘"

	return SeparatorStyle.Render(top) + "\n" +
		SeparatorStyle.Render("│") + titleContent + repeatChar(" ", titlePadding) + SeparatorStyle.Render("│") + "\n" +
		SeparatorStyle.Render("│") + subtitleContent + repeatChar(" ", subtitlePadding) + SeparatorStyle.Render("│") + "\n" +
		SeparatorStyle.Render(bottom)
}

// repeatChar repeats a character n times
func repeatChar(char string, n int) string {
	if n <= 0 {
		return ""
	}
	result := ""
	for i := 0; i < n; i++ {
		result += char
	}
	return result
}

// GetStatusIndicator returns the status indicator with text label
func GetStatusIndicator(running bool, unknown bool) string {
	if unknown {
		return StatusUnknown.Render("? unknown")
	}
	if running {
		return StatusRunning.Render("● running")
	}
	return StatusStopped.Render("○ stopped")
}
