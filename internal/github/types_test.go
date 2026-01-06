package github

import (
	"encoding/json"
	"testing"
)

func TestIssueState_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected IssueState
	}{
		{"uppercase OPEN", `"OPEN"`, IssueStateOpen},
		{"lowercase open", `"open"`, IssueStateOpen},
		{"mixed case Open", `"Open"`, IssueStateOpen},
		{"uppercase CLOSED", `"CLOSED"`, IssueStateClosed},
		{"lowercase closed", `"closed"`, IssueStateClosed},
		{"mixed case Closed", `"Closed"`, IssueStateClosed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var state IssueState
			if err := json.Unmarshal([]byte(tt.input), &state); err != nil {
				t.Fatalf("UnmarshalJSON failed: %v", err)
			}
			if state != tt.expected {
				t.Errorf("got %q, want %q", state, tt.expected)
			}
		})
	}
}

func TestIssueState_UnmarshalJSON_ErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"integer value", `123`},
		{"boolean value", `true`},
		{"object value", `{"state": "open"}`},
		{"array value", `["open"]`},
		{"malformed JSON", `"open`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var state IssueState
			err := json.Unmarshal([]byte(tt.input), &state)
			if err == nil {
				t.Errorf("expected error for input %s, got nil", tt.input)
			}
		})
	}
}

func TestIssueState_UnmarshalJSON_NullValue(t *testing.T) {
	// null is valid JSON and results in empty string (handled gracefully)
	var state IssueState
	err := json.Unmarshal([]byte(`null`), &state)
	if err != nil {
		t.Errorf("unexpected error for null: %v", err)
	}
	if state != "" {
		t.Errorf("expected empty string for null, got %q", state)
	}
}

func TestIssue_UnmarshalJSON(t *testing.T) {
	// Test full Issue struct unmarshaling with uppercase state
	input := `{"number": 42, "title": "Test Issue", "state": "OPEN", "url": "https://example.com"}`

	var issue Issue
	if err := json.Unmarshal([]byte(input), &issue); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	if issue.State != IssueStateOpen {
		t.Errorf("State = %q, want %q", issue.State, IssueStateOpen)
	}
	if issue.Number != 42 {
		t.Errorf("Number = %d, want 42", issue.Number)
	}
	if issue.Title != "Test Issue" {
		t.Errorf("Title = %q, want %q", issue.Title, "Test Issue")
	}
}

func TestIssue_UnmarshalJSON_ClosedState(t *testing.T) {
	input := `{"number": 99, "title": "Closed Issue", "state": "CLOSED", "url": "https://example.com"}`

	var issue Issue
	if err := json.Unmarshal([]byte(input), &issue); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	if issue.State != IssueStateClosed {
		t.Errorf("State = %q, want %q", issue.State, IssueStateClosed)
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func TestConfig_IsAutoLabelEnabled(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected bool
	}{
		{"nil value defaults to true", Config{}, true},
		{"explicit true", Config{AutoLabelIssues: boolPtr(true)}, true},
		{"explicit false", Config{AutoLabelIssues: boolPtr(false)}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.IsAutoLabelEnabled(); got != tt.expected {
				t.Errorf("IsAutoLabelEnabled() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestConfig_ShouldCreateLabelIfMissing(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected bool
	}{
		{"nil value defaults to true", Config{}, true},
		{"explicit true", Config{CreateLabelIfMissing: boolPtr(true)}, true},
		{"explicit false", Config{CreateLabelIfMissing: boolPtr(false)}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.ShouldCreateLabelIfMissing(); got != tt.expected {
				t.Errorf("ShouldCreateLabelIfMissing() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.DefaultState != IssueStateOpen {
		t.Errorf("DefaultState = %q, want %q", cfg.DefaultState, IssueStateOpen)
	}
	if cfg.BranchPrefix != "issue-" {
		t.Errorf("BranchPrefix = %q, want %q", cfg.BranchPrefix, "issue-")
	}
	if cfg.MaxIssues != 50 {
		t.Errorf("MaxIssues = %d, want 50", cfg.MaxIssues)
	}
	if cfg.InProgressLabel != "in-progress" {
		t.Errorf("InProgressLabel = %q, want %q", cfg.InProgressLabel, "in-progress")
	}
}
