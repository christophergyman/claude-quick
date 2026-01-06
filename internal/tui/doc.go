// Package tui implements the terminal user interface using Bubble Tea.
//
// # Architecture
//
// The TUI operates as a deterministic state machine. The Model struct holds
// all application state, and the Update method handles all state transitions
// in response to messages.
//
// # State Machine
//
// Key states include:
//   - StateDashboard: Main container list view
//   - StateContainerStarting/Stopping: Container operations in progress
//   - StateTmuxSelect: Session picker after container starts
//   - StateNewWorktreeInput: Creating git worktree
//   - StateGitHubIssueSelect: GitHub issue selection for worktree names
//
// All transitions are explicit via message handling in Update().
//
// # Async Command Pattern
//
// No blocking I/O in the UI. Operations return tea.Cmd that execute async:
//
//	discoverInstances() → instancesDiscoveredMsg
//	startContainer()    → containerStartedMsg
//	loadTmuxSessions()  → tmuxSessionsLoadedMsg
//
// # Key Files
//
//   - model.go: Model definition and Update/View methods
//   - state.go: State constants and transitions
//   - handlers.go: Keyboard event handlers
//   - commands.go: Async command implementations
//   - messages.go: Message types for async results
//   - container.go: Dashboard rendering
//   - tmux.go: Session selection rendering
//   - styles.go: Lipgloss styling
package tui
