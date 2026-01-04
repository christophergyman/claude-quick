package tui

import "github.com/charmbracelet/lipgloss"

// Colors - high contrast for mobile visibility
var (
	colorPrimary   = lipgloss.Color("#7C3AED") // Purple
	colorSecondary = lipgloss.Color("#A78BFA") // Light purple
	colorMuted     = lipgloss.Color("#6B7280") // Gray
	colorSuccess   = lipgloss.Color("#10B981") // Green
	colorWarning   = lipgloss.Color("#F59E0B") // Amber
	colorError     = lipgloss.Color("#EF4444") // Red
	colorWhite     = lipgloss.Color("#FFFFFF")
)

// Styles
var (
	// Title style for headers
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			MarginBottom(1)

	// Subtitle style
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			MarginBottom(1)

	// Selected item style - bold with indicator
	SelectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite).
			Background(colorPrimary).
			Padding(0, 1)

	// Unselected item style
	ItemStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Padding(0, 1)

	// Dimmed item style (for paths, details)
	DimmedStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	// Help text style
	HelpStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			MarginTop(1)

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
			BorderForeground(colorPrimary).
			Padding(1, 2)

	// Input style for text input
	InputStyle = lipgloss.NewStyle().
			Foreground(colorWhite).
			Background(colorPrimary).
			Padding(0, 1)

	// Spinner style
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(colorPrimary)
)

// Cursor returns the selection cursor
func Cursor() string {
	return lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true).
		Render("> ")
}

// NoCursor returns spacing for non-selected items
func NoCursor() string {
	return "  "
}
