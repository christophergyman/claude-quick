// Package constants provides centralized configuration constants for claude-quick.
// This eliminates magic numbers scattered across the codebase.
package constants

// Container timeout constants (in seconds)
const (
	DefaultContainerTimeout = 300  // Default timeout for container operations
	MinContainerTimeout     = 30   // Minimum allowed timeout
	MaxContainerTimeout     = 1800 // Maximum allowed timeout (30 minutes)
)

// Discovery constants
const (
	DefaultMaxDepth = 3 // Default directory search depth
)

// Text input UI constants
const (
	TextInputCharLimit = 50 // Character limit for text input fields
	TextInputWidth     = 30 // Width of text input fields in characters
)

// Display constants
const (
	SHATruncateLength      = 7  // Length for truncated git SHA display
	DefaultPathTruncateLen = 40 // Default max length for path display
	PathTruncatePadding    = 6  // Padding to subtract from width for path display
)

// Devcontainer file and directory names
const (
	DevcontainerDir        = ".devcontainer"
	DevcontainerConfigFile = "devcontainer.json"
)

// Reserved branch names that cannot be used for worktrees
var ReservedBranchNames = []string{"main", "master"}

// IsReservedBranchName checks if a branch name is reserved
func IsReservedBranchName(name string) bool {
	for _, reserved := range ReservedBranchNames {
		if name == reserved {
			return true
		}
	}
	return false
}

// Default values for configuration
const (
	DefaultSessionName       = "main"
	DefaultWorktreePlaceholder = "feature-branch"
	DefaultBranchUnknown     = "unknown"
)

// DefaultExcludedDirs returns the default directories to exclude from scanning
func DefaultExcludedDirs() []string {
	return []string{
		"node_modules",
		"vendor",
		".git",
		"__pycache__",
		"venv",
		".venv",
		"dist",
		"build",
		".cache",
	}
}
