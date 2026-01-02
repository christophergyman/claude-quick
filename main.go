package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/chezu/claude-quick/internal/config"
	"github.com/chezu/claude-quick/internal/devcontainer"
	"github.com/chezu/claude-quick/internal/tui"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Check for devcontainer CLI
	if err := devcontainer.CheckCLI(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Discover devcontainer projects
	projects := devcontainer.Discover(cfg.SearchPaths, cfg.MaxDepth)

	// Create and run TUI
	model := tui.New(projects)
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}

	// After TUI exits, check if we should attach to a tmux session
	if m, ok := finalModel.(tui.Model); ok {
		projectPath, sessionName, shouldAttach := m.GetAttachInfo()
		if shouldAttach {
			// Execute into container and attach to tmux
			// This replaces the current process
			err := devcontainer.ExecInteractive(projectPath, []string{"tmux", "attach", "-t", sessionName})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error attaching to tmux: %v\n", err)
				os.Exit(1)
			}
		}
	}
}
