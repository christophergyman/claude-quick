// Package github provides GitHub integration for fetching issues and generating branch names.
package github

// IssueState represents the state of a GitHub issue.
type IssueState string

const (
	// IssueStateOpen represents open issues.
	IssueStateOpen IssueState = "open"
	// IssueStateClosed represents closed issues.
	IssueStateClosed IssueState = "closed"
	// IssueStateAll represents all issues regardless of state.
	IssueStateAll IssueState = "all"
)

// Issue represents a GitHub issue.
type Issue struct {
	Number int        `json:"number"`
	Title  string     `json:"title"`
	State  IssueState `json:"state"`
	URL    string     `json:"url"`
	Body   string     `json:"body"`
}

// Config holds GitHub-related configuration.
type Config struct {
	DefaultState IssueState `yaml:"default_state"`
	BranchPrefix string     `yaml:"branch_prefix"`
	MaxIssues    int        `yaml:"max_issues"`
}

// DefaultConfig returns the default GitHub configuration.
func DefaultConfig() Config {
	return Config{
		DefaultState: IssueStateOpen,
		BranchPrefix: "issue-",
		MaxIssues:    50,
	}
}
