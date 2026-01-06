package tui

import "github.com/charmbracelet/lipgloss"

// colorPalette holds all colors for a theme
type colorPalette struct {
	orange    lipgloss.Color
	primary   lipgloss.Color
	dim       lipgloss.Color
	success   lipgloss.Color
	warning   lipgloss.Color
	errorCol  lipgloss.Color
	separator lipgloss.Color
}

// Dark mode palette (for dark terminal backgrounds)
var darkPalette = colorPalette{
	orange:    lipgloss.Color("#E07A5F"), // Anthropic orange (terracotta/salmon)
	primary:   lipgloss.Color("#FFFFFF"), // White for primary text
	dim:       lipgloss.Color("#6B7280"), // Gray for dimmed text
	success:   lipgloss.Color("#10B981"), // Green for running
	warning:   lipgloss.Color("#F59E0B"), // Yellow/amber for stopped
	errorCol:  lipgloss.Color("#EF4444"), // Red for errors
	separator: lipgloss.Color("#4B5563"), // Darker gray for separators
}

// Light mode palette (for light terminal backgrounds)
var lightPalette = colorPalette{
	orange:    lipgloss.Color("#C45A3E"), // Darker orange for contrast
	primary:   lipgloss.Color("#1F2937"), // Dark gray for primary text
	dim:       lipgloss.Color("#6B7280"), // Medium gray
	success:   lipgloss.Color("#059669"), // Darker green
	warning:   lipgloss.Color("#D97706"), // Darker amber
	errorCol:  lipgloss.Color("#DC2626"), // Darker red
	separator: lipgloss.Color("#9CA3AF"), // Light gray borders
}

// currentPalette holds the active color palette
var currentPalette = darkPalette

// IsDarkMode tracks the current theme state
var IsDarkMode = true

// Styles - initialized with dark mode colors
var (
	// Header title style (Anthropic orange)
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(currentPalette.orange)

	// Subtitle style
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(currentPalette.dim)

	// Selected item style - bold primary color (no background)
	SelectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(currentPalette.primary)

	// Unselected item style
	ItemStyle = lipgloss.NewStyle().
			Foreground(currentPalette.primary)

	// Dimmed item style (for paths, details)
	DimmedStyle = lipgloss.NewStyle().
			Foreground(currentPalette.dim)

	// Help text style
	HelpStyle = lipgloss.NewStyle().
			Foreground(currentPalette.dim)

	// Error style
	ErrorStyle = lipgloss.NewStyle().
			Foreground(currentPalette.errorCol).
			Bold(true)

	// Success style
	SuccessStyle = lipgloss.NewStyle().
			Foreground(currentPalette.success)

	// Warning style
	WarningStyle = lipgloss.NewStyle().
			Foreground(currentPalette.warning)

	// Box style for containers
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(currentPalette.dim).
			Padding(0, 1)

	// Input style for text input
	InputStyle = lipgloss.NewStyle().
			Foreground(currentPalette.primary)

	// Spinner style
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(currentPalette.orange)

	// Separator style
	SeparatorStyle = lipgloss.NewStyle().
			Foreground(currentPalette.separator)

	// Key highlight style (for keybinding display)
	KeyStyle = lipgloss.NewStyle().
			Foreground(currentPalette.primary)

	// Column header style
	ColumnHeaderStyle = lipgloss.NewStyle().
				Foreground(currentPalette.dim).
				Bold(true)
)

// Status text styles with labels
var (
	StatusRunning    = lipgloss.NewStyle().Foreground(currentPalette.success)
	StatusStopped    = lipgloss.NewStyle().Foreground(currentPalette.warning)
	StatusUnknown    = lipgloss.NewStyle().Foreground(currentPalette.dim)
	StatusInProgress = lipgloss.NewStyle().Foreground(currentPalette.orange)
)

// ApplyTheme updates all styles based on the dark mode setting
func ApplyTheme(darkMode bool) {
	IsDarkMode = darkMode
	if darkMode {
		currentPalette = darkPalette
	} else {
		currentPalette = lightPalette
	}

	// Rebuild all styles with the new palette
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(currentPalette.orange)

	SubtitleStyle = lipgloss.NewStyle().
		Foreground(currentPalette.dim)

	SelectedStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(currentPalette.primary)

	ItemStyle = lipgloss.NewStyle().
		Foreground(currentPalette.primary)

	DimmedStyle = lipgloss.NewStyle().
		Foreground(currentPalette.dim)

	HelpStyle = lipgloss.NewStyle().
		Foreground(currentPalette.dim)

	ErrorStyle = lipgloss.NewStyle().
		Foreground(currentPalette.errorCol).
		Bold(true)

	SuccessStyle = lipgloss.NewStyle().
		Foreground(currentPalette.success)

	WarningStyle = lipgloss.NewStyle().
		Foreground(currentPalette.warning)

	BoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(currentPalette.dim).
		Padding(0, 1)

	InputStyle = lipgloss.NewStyle().
		Foreground(currentPalette.primary)

	SpinnerStyle = lipgloss.NewStyle().
		Foreground(currentPalette.orange)

	SeparatorStyle = lipgloss.NewStyle().
		Foreground(currentPalette.separator)

	KeyStyle = lipgloss.NewStyle().
		Foreground(currentPalette.primary)

	ColumnHeaderStyle = lipgloss.NewStyle().
		Foreground(currentPalette.dim).
		Bold(true)

	// Status styles
	StatusRunning = lipgloss.NewStyle().Foreground(currentPalette.success)
	StatusStopped = lipgloss.NewStyle().Foreground(currentPalette.warning)
	StatusUnknown = lipgloss.NewStyle().Foreground(currentPalette.dim)
	StatusInProgress = lipgloss.NewStyle().Foreground(currentPalette.orange)
}

// Cursor returns the selection cursor (› instead of >)
func Cursor() string {
	return lipgloss.NewStyle().
		Foreground(currentPalette.orange).
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
