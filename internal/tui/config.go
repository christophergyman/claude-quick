package tui

import (
	"fmt"

	"github.com/christophergyman/claude-quick/internal/config"
)

// RenderConfigDisplay renders the configuration view
func RenderConfigDisplay(cfg *config.Config) string {
	b := renderWithHeader("Configuration")

	// Config file location
	b.WriteString(DimmedStyle.Render("Config file: "))
	b.WriteString(config.ConfigPath())
	b.WriteString("\n\n")

	// Section separator
	b.WriteString("  " + RenderSeparator(defaultWidth-4))
	b.WriteString("\n\n")

	// Search Paths
	b.WriteString(ColumnHeaderStyle.Render("Search Paths"))
	b.WriteString("\n")
	for _, p := range cfg.SearchPaths {
		b.WriteString("  " + p + "\n")
	}
	b.WriteString("\n")

	// Max Depth
	b.WriteString(ColumnHeaderStyle.Render("Max Depth: "))
	b.WriteString(fmt.Sprintf("%d", cfg.MaxDepth))
	b.WriteString("\n\n")

	// Excluded Dirs (show all)
	b.WriteString(ColumnHeaderStyle.Render("Excluded Dirs"))
	b.WriteString("\n")
	for _, d := range cfg.ExcludedDirs {
		b.WriteString("  " + DimmedStyle.Render(d) + "\n")
	}
	b.WriteString("\n")

	// Default Session Name
	b.WriteString(ColumnHeaderStyle.Render("Default Session: "))
	b.WriteString(cfg.DefaultSessionName)
	b.WriteString("\n\n")

	// Container Timeout
	b.WriteString(ColumnHeaderStyle.Render("Container Timeout: "))
	b.WriteString(fmt.Sprintf("%ds", cfg.ContainerTimeout))
	b.WriteString("\n\n")

	// Footer
	b.WriteString("  " + RenderSeparator(defaultWidth-4))
	b.WriteString("\n")
	b.WriteString(RenderKeyBinding("any key", "return"))

	return b.String()
}
