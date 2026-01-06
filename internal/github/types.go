// Package github provides GitHub integration for fetching issues and generating branch names.
package github

import (
	"encoding/json"
	"strings"
)

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

// UnmarshalJSON normalizes the state value to lowercase.
// GitHub API returns uppercase ("OPEN", "CLOSED") but we use lowercase constants.
func (s *IssueState) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*s = IssueState(strings.ToLower(raw))
	return nil
}

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
	DefaultState         IssueState `yaml:"default_state"`
	BranchPrefix         string     `yaml:"branch_prefix"`
	MaxIssues            int        `yaml:"max_issues"`
	InProgressLabel      string     `yaml:"in_progress_label,omitempty"`
	LabelColor           string     `yaml:"label_color,omitempty"`
	LabelDescription     string     `yaml:"label_description,omitempty"`
	AutoLabelIssues      *bool      `yaml:"auto_label_issues,omitempty"`
	CreateLabelIfMissing *bool      `yaml:"create_label_if_missing,omitempty"`
}

// IsAutoLabelEnabled returns whether to auto-label issues on worktree creation.
func (c Config) IsAutoLabelEnabled() bool {
	if c.AutoLabelIssues == nil {
		return true // Default enabled
	}
	return *c.AutoLabelIssues
}

// ShouldCreateLabelIfMissing returns whether to create the label if it doesn't exist.
func (c Config) ShouldCreateLabelIfMissing() bool {
	if c.CreateLabelIfMissing == nil {
		return true // Default enabled
	}
	return *c.CreateLabelIfMissing
}

// DefaultConfig returns the default GitHub configuration.
func DefaultConfig() Config {
	return Config{
		DefaultState:     IssueStateOpen,
		BranchPrefix:     "issue-",
		MaxIssues:        50,
		InProgressLabel:  "in-progress",
		LabelColor:       "fbca04",
		LabelDescription: "Issue is being actively worked on",
	}
}
