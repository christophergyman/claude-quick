package tui

// State represents the current view state of the TUI.
// The application is a state machine that transitions between these states
// based on user input and async operation results.
type State int

const (
	// StateDiscovering is the initial state while scanning for devcontainer projects
	StateDiscovering State = iota
	// StateRefreshingStatus is shown while fetching container status from Docker
	StateRefreshingStatus
	// StateDashboard is the main view showing all discovered containers
	StateDashboard
	// StateContainerStarting is shown while a container is being started
	StateContainerStarting
	// StateConfirmStop prompts user to confirm stopping a container
	StateConfirmStop
	// StateConfirmRestart prompts user to confirm restarting a container
	StateConfirmRestart
	// StateContainerStopping is shown while a container is being stopped
	StateContainerStopping
	// StateContainerRestarting is shown while a container is being restarted
	StateContainerRestarting
	// StateTmuxSelect shows the list of tmux sessions in a container
	StateTmuxSelect
	// StateNewSessionInput shows text input for new session name
	StateNewSessionInput
	// StateAttaching is shown while attaching to a tmux session
	StateAttaching
	// StateConfirmTmuxStop prompts user to confirm killing a tmux session
	StateConfirmTmuxStop
	// StateConfirmTmuxRestart prompts user to confirm restarting a tmux session
	StateConfirmTmuxRestart
	// StateTmuxStopping is shown while a tmux session is being killed
	StateTmuxStopping
	// StateTmuxRestarting is shown while a tmux session is being restarted
	StateTmuxRestarting
	// StateLoadingTmuxSessions is shown while loading tmux sessions from a container
	StateLoadingTmuxSessions
	// StateError displays an error message
	StateError
	// StateShowConfig displays current configuration
	StateShowConfig
	// StateNewWorktreeInput shows text input for new worktree branch name
	StateNewWorktreeInput
	// StateCreatingWorktree is shown while creating a new git worktree
	StateCreatingWorktree
	// StateConfirmDeleteWorktree prompts user to confirm deleting a worktree
	StateConfirmDeleteWorktree
	// StateDeletingWorktree is shown while deleting a git worktree
	StateDeletingWorktree
	// StateGitHubIssuesLoading is shown while fetching issues from GitHub
	StateGitHubIssuesLoading
	// StateGitHubIssuesList displays the list of GitHub issues
	StateGitHubIssuesList
	// StateGitHubIssueDetailLoading is shown while loading issue details
	StateGitHubIssueDetailLoading
	// StateGitHubIssueDetail displays a single issue's details
	StateGitHubIssueDetail
	// StateGitHubWorktreeCreating is shown while creating worktree from issue
	StateGitHubWorktreeCreating

	// Wizard states for guided configuration setup
	// StateWizardWelcome is the introduction screen for the setup wizard
	StateWizardWelcome
	// StateWizardSearchPaths is where users configure directories to scan
	StateWizardSearchPaths
	// StateWizardCredentials is where users configure optional credentials (GITHUB_TOKEN)
	StateWizardCredentials
	// StateWizardSettings is where users configure default settings
	StateWizardSettings
	// StateWizardSummary shows the configuration summary before saving
	StateWizardSummary
	// StateWizardSaving is shown while saving the configuration
	StateWizardSaving
)
