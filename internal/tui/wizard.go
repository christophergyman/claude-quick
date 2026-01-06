package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/christophergyman/claude-quick/internal/auth"
)

// Wizard step constants
const (
	WizardStepWelcome = iota
	WizardStepSearchPaths
	WizardStepCredentials
	WizardStepSettings
	WizardStepSummary
	WizardStepCount = 5
)

// wizardStepNames provides labels for the progress indicator
var wizardStepNames = []string{
	"Welcome",
	"Paths",
	"Credentials",
	"Settings",
	"Summary",
}

// renderWizardProgress renders the step progress indicator
func renderWizardProgress(currentStep int, width int) string {
	if width <= 0 {
		width = defaultWidth
	}

	var b strings.Builder

	// Step X of Y header
	b.WriteString(DimmedStyle.Render(fmt.Sprintf("Step %d of %d", currentStep+1, WizardStepCount)))
	b.WriteString("\n")

	// Progress dots
	for i := 0; i < WizardStepCount; i++ {
		if i == currentStep {
			b.WriteString(SuccessStyle.Render("[●] "))
			b.WriteString(SelectedStyle.Render(wizardStepNames[i]))
		} else if i < currentStep {
			b.WriteString(SuccessStyle.Render("[●] "))
			b.WriteString(DimmedStyle.Render(wizardStepNames[i]))
		} else {
			b.WriteString(DimmedStyle.Render("[○] "))
			b.WriteString(DimmedStyle.Render(wizardStepNames[i]))
		}
		if i < WizardStepCount-1 {
			b.WriteString("  ")
		}
	}
	b.WriteString("\n")

	return b.String()
}

// RenderWizardWelcome renders the welcome/introduction screen
func RenderWizardWelcome(width int) string {
	if width <= 0 {
		width = defaultWidth
	}

	var b strings.Builder

	// Header
	b.WriteString(RenderBorderedHeader("claude-quick", "Configuration Wizard", width))
	b.WriteString("\n\n")

	// Progress
	b.WriteString(renderWizardProgress(WizardStepWelcome, width))
	b.WriteString("\n")

	// Welcome message
	b.WriteString(SelectedStyle.Render("Welcome to claude-quick!"))
	b.WriteString("\n\n")

	b.WriteString("This wizard will help you set up your configuration:\n\n")

	b.WriteString("  ")
	b.WriteString(SuccessStyle.Render("1."))
	b.WriteString(" Configure search paths for devcontainer projects\n")

	b.WriteString("  ")
	b.WriteString(SuccessStyle.Render("2."))
	b.WriteString(" Set up credentials (optional)\n")

	b.WriteString("  ")
	b.WriteString(SuccessStyle.Render("3."))
	b.WriteString(" Configure default settings\n")

	b.WriteString("  ")
	b.WriteString(SuccessStyle.Render("4."))
	b.WriteString(" Review and save\n")

	b.WriteString("\n")
	b.WriteString(DimmedStyle.Render("Your configuration will be saved next to the executable."))
	b.WriteString("\n\n")

	// Footer
	b.WriteString(RenderSeparator(width - 4))
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("  enter next  q/esc cancel"))

	return b.String()
}

// RenderWizardSearchPaths renders the search paths configuration screen
func RenderWizardSearchPaths(paths []string, warnings map[string]bool, cursor int, input interface{ View() string }, editMode bool, width int) string {
	if width <= 0 {
		width = defaultWidth
	}

	var b strings.Builder

	// Header
	b.WriteString(RenderBorderedHeader("claude-quick", "Configuration Wizard", width))
	b.WriteString("\n\n")

	// Progress
	b.WriteString(renderWizardProgress(WizardStepSearchPaths, width))
	b.WriteString("\n")

	// Title
	b.WriteString(SelectedStyle.Render("Search Paths"))
	b.WriteString("\n")
	b.WriteString(DimmedStyle.Render("Directories to scan for devcontainer projects"))
	b.WriteString("\n\n")

	if editMode {
		// Show input field for new path
		b.WriteString("Enter path (~ will be expanded):\n\n")
		b.WriteString(input.View())
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render("  enter add  esc cancel"))
	} else {
		// Show list of paths
		if len(paths) == 0 {
			b.WriteString(WarningStyle.Render("  No paths configured. Add at least one search path."))
			b.WriteString("\n")
		} else {
			for i, path := range paths {
				// Check if path exists
				exists, checked := warnings[path]
				var pathStyle lipgloss.Style
				var indicator string

				if checked && !exists {
					pathStyle = WarningStyle
					indicator = " (not found)"
				} else if checked {
					pathStyle = SuccessStyle
					indicator = ""
				} else {
					pathStyle = ItemStyle
					indicator = ""
				}

				if i == cursor {
					b.WriteString(Cursor())
					b.WriteString(pathStyle.Render(path))
					b.WriteString(WarningStyle.Render(indicator))
				} else {
					b.WriteString(NoCursor())
					b.WriteString(pathStyle.Render(path))
					b.WriteString(WarningStyle.Render(indicator))
				}
				b.WriteString("\n")
			}
		}
		b.WriteString("\n")

		// Footer
		b.WriteString(RenderSeparator(width - 4))
		b.WriteString("\n")
		b.WriteString(HelpStyle.Render("  a add  d delete  ↑↓ navigate  enter/tab next  backspace back"))
	}

	return b.String()
}

// RenderWizardCredentials renders the credentials setup screen
func RenderWizardCredentials(credentials []auth.Credential, sourceType auth.SourceType, valueInput interface{ View() string }, cursor int, editMode bool, width int) string {
	if width <= 0 {
		width = defaultWidth
	}

	var b strings.Builder

	// Header
	b.WriteString(RenderBorderedHeader("claude-quick", "Configuration Wizard", width))
	b.WriteString("\n\n")

	// Progress
	b.WriteString(renderWizardProgress(WizardStepCredentials, width))
	b.WriteString("\n")

	// Title
	b.WriteString(SelectedStyle.Render("Credentials"))
	b.WriteString(" ")
	b.WriteString(DimmedStyle.Render("(optional)"))
	b.WriteString("\n")
	b.WriteString(DimmedStyle.Render("Configure GITHUB_TOKEN for git operations"))
	b.WriteString("\n\n")

	// Check if GITHUB_TOKEN already exists
	hasGithubToken := false
	for _, c := range credentials {
		if c.Name == "GITHUB_TOKEN" {
			hasGithubToken = true
			break
		}
	}

	if editMode {
		// Show source type selection and value input
		if hasGithubToken {
			b.WriteString(WarningStyle.Render("GITHUB_TOKEN already configured. Press esc to cancel."))
			b.WriteString("\n\n")
		}
		b.WriteString("Select source type:\n\n")

		sources := []auth.SourceType{auth.SourceFile, auth.SourceEnv, auth.SourceCommand}
		sourceLabels := map[auth.SourceType]string{
			auth.SourceFile:    "file    - Read from a file (e.g., ~/.github_token)",
			auth.SourceEnv:     "env     - Read from environment variable",
			auth.SourceCommand: "command - Run command to get value (e.g., 1Password CLI)",
		}

		for _, src := range sources {
			if src == sourceType {
				b.WriteString(Cursor())
				b.WriteString(SelectedStyle.Render(sourceLabels[src]))
			} else {
				b.WriteString(NoCursor())
				b.WriteString(ItemStyle.Render(sourceLabels[src]))
			}
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString("Value:\n")
		b.WriteString(valueInput.View())
		b.WriteString("\n\n")

		// Help text based on source type
		switch sourceType {
		case auth.SourceFile:
			b.WriteString(DimmedStyle.Render("Enter path to file containing the token"))
		case auth.SourceEnv:
			b.WriteString(DimmedStyle.Render("Enter environment variable name"))
		case auth.SourceCommand:
			b.WriteString(DimmedStyle.Render("Enter command to execute"))
		}
		b.WriteString("\n\n")

		b.WriteString(HelpStyle.Render("  ↑↓ change source  enter add  esc cancel"))
	} else {
		// Show configured credentials
		if len(credentials) == 0 {
			b.WriteString(DimmedStyle.Render("  No credentials configured."))
			b.WriteString("\n")
		} else {
			for i, cred := range credentials {
				if i == cursor {
					b.WriteString(Cursor())
					b.WriteString(SelectedStyle.Render(cred.Name))
				} else {
					b.WriteString(NoCursor())
					b.WriteString(ItemStyle.Render(cred.Name))
				}
				b.WriteString(" ")
				b.WriteString(DimmedStyle.Render(fmt.Sprintf("(%s: %s)", cred.Source, cred.Value)))
				b.WriteString("\n")
			}
		}
		b.WriteString("\n")

		// Footer
		b.WriteString(RenderSeparator(width - 4))
		b.WriteString("\n")
		b.WriteString(HelpStyle.Render("  a add  d delete  s skip  enter/tab next  backspace back"))
	}

	return b.String()
}

// RenderWizardSettings renders the general settings screen
func RenderWizardSettings(sessionName, timeout, launchCmd, maxDepth string, darkMode bool, cursor int, editMode bool, activeInput interface{ View() string }, width int) string {
	if width <= 0 {
		width = defaultWidth
	}

	var b strings.Builder

	// Header
	b.WriteString(RenderBorderedHeader("claude-quick", "Configuration Wizard", width))
	b.WriteString("\n\n")

	// Progress
	b.WriteString(renderWizardProgress(WizardStepSettings, width))
	b.WriteString("\n")

	// Title
	b.WriteString(SelectedStyle.Render("Default Settings"))
	b.WriteString("\n\n")

	// Settings fields (5 items: session name, timeout, launch cmd, max depth, dark mode)
	settingsLabels := []string{
		"Default session name",
		"Container timeout (seconds)",
		"Launch command",
		"Max search depth",
		"Dark mode",
	}
	settingsValues := []string{sessionName, timeout, launchCmd, maxDepth, ""}
	settingsHints := []string{
		"Name for new tmux sessions",
		"Timeout for container operations",
		"Command to run in new sessions",
		"How deep to search for projects",
		"Toggle terminal color scheme",
	}

	// Check for invalid numeric values
	_, timeoutErr := strconv.Atoi(timeout)
	_, maxDepthErr := strconv.Atoi(maxDepth)
	invalidFields := map[int]bool{
		1: timeout != "" && timeoutErr != nil,
		3: maxDepth != "" && maxDepthErr != nil,
	}

	for i, label := range settingsLabels {
		if i == cursor {
			b.WriteString(Cursor())
			b.WriteString(SelectedStyle.Render(label + ": "))
		} else {
			b.WriteString(NoCursor())
			b.WriteString(ItemStyle.Render(label + ": "))
		}

		if i == 4 {
			// Dark mode toggle
			if darkMode {
				b.WriteString(SuccessStyle.Render("[x] enabled"))
			} else {
				b.WriteString(DimmedStyle.Render("[ ] disabled"))
			}
		} else if editMode && i == cursor {
			// Show input for active field
			b.WriteString("\n    ")
			b.WriteString(activeInput.View())
		} else {
			// Show value with warning if invalid
			if invalidFields[i] {
				b.WriteString(WarningStyle.Render(settingsValues[i]))
				b.WriteString(WarningStyle.Render(" (invalid, will use default)"))
			} else {
				b.WriteString(SuccessStyle.Render(settingsValues[i]))
			}
		}
		b.WriteString("\n")

		// Show hint for current selection
		if i == cursor && !editMode {
			b.WriteString("    ")
			b.WriteString(DimmedStyle.Render(settingsHints[i]))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	// Footer
	b.WriteString(RenderSeparator(width - 4))
	b.WriteString("\n")
	if editMode {
		b.WriteString(HelpStyle.Render("  enter confirm  esc cancel"))
	} else {
		b.WriteString(HelpStyle.Render("  ↑↓ navigate  enter edit  t toggle dark mode  tab next  backspace back"))
	}

	return b.String()
}

// RenderWizardSummary renders the configuration summary screen
func RenderWizardSummary(paths []string, credentials []auth.Credential, sessionName, timeout, launchCmd, maxDepth string, darkMode bool, configPath string, width int) string {
	if width <= 0 {
		width = defaultWidth
	}

	var b strings.Builder

	// Header
	b.WriteString(RenderBorderedHeader("claude-quick", "Configuration Wizard", width))
	b.WriteString("\n\n")

	// Progress
	b.WriteString(renderWizardProgress(WizardStepSummary, width))
	b.WriteString("\n")

	// Title
	b.WriteString(SelectedStyle.Render("Configuration Summary"))
	b.WriteString("\n\n")

	// Search paths section
	b.WriteString(ColumnHeaderStyle.Render("SEARCH PATHS"))
	b.WriteString("\n")
	if len(paths) == 0 {
		b.WriteString("  ")
		b.WriteString(DimmedStyle.Render("(none)"))
		b.WriteString("\n")
	} else {
		for _, path := range paths {
			b.WriteString("  ")
			b.WriteString(ItemStyle.Render(path))
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")

	// Credentials section
	b.WriteString(ColumnHeaderStyle.Render("CREDENTIALS"))
	b.WriteString("\n")
	if len(credentials) == 0 {
		b.WriteString("  ")
		b.WriteString(DimmedStyle.Render("(none)"))
		b.WriteString("\n")
	} else {
		for _, cred := range credentials {
			b.WriteString("  ")
			b.WriteString(ItemStyle.Render(cred.Name))
			b.WriteString(" ")
			b.WriteString(DimmedStyle.Render(fmt.Sprintf("(%s: %s)", cred.Source, cred.Value)))
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")

	// Settings section
	b.WriteString(ColumnHeaderStyle.Render("SETTINGS"))
	b.WriteString("\n")

	// Show validated/defaulted values (what will actually be saved)
	displaySession := sessionName
	if displaySession == "" {
		displaySession = "(default: main)"
	}
	b.WriteString(fmt.Sprintf("  Default Session:    %s\n", SuccessStyle.Render(displaySession)))

	b.WriteString(fmt.Sprintf("  Container Timeout:  %s\n", SuccessStyle.Render(timeout+"s")))

	displayLaunch := launchCmd
	if displayLaunch == "" {
		displayLaunch = "(none)"
	}
	b.WriteString(fmt.Sprintf("  Launch Command:     %s\n", SuccessStyle.Render(displayLaunch)))

	b.WriteString(fmt.Sprintf("  Max Depth:          %s\n", SuccessStyle.Render(maxDepth)))
	darkModeStr := "enabled"
	if !darkMode {
		darkModeStr = "disabled"
	}
	b.WriteString(fmt.Sprintf("  Dark Mode:          %s\n", SuccessStyle.Render(darkModeStr)))
	b.WriteString("\n")

	// Config path
	b.WriteString(RenderSeparator(width - 4))
	b.WriteString("\n")
	b.WriteString(DimmedStyle.Render("Config will be saved to:"))
	b.WriteString("\n")
	b.WriteString(SuccessStyle.Render(configPath))
	b.WriteString("\n\n")

	// Footer
	b.WriteString(HelpStyle.Render("  enter save  backspace edit  q cancel"))

	return b.String()
}

// RenderWizardSaving renders the saving progress screen
func RenderWizardSaving(spinnerView string) string {
	b := renderSimpleHeader("Configuration Wizard")
	b.WriteString(SpinnerStyle.Render(spinnerView))
	b.WriteString(" Saving configuration...")
	return b.String()
}
