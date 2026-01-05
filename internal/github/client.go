package github

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// CheckCLI verifies that gh CLI is installed and authenticated.
func CheckCLI() error {
	// First check if gh is installed
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("GitHub CLI not found. Install with: brew install gh")
	}

	// Check if authenticated
	cmd := exec.Command("gh", "auth", "status")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("not authenticated with GitHub CLI. Run: gh auth login")
	}
	return nil
}

// DetectRepository determines the GitHub owner/repo from git remote.
func DetectRepository(repoPath string) (owner, repo string, err error) {
	// Run: git -C <path> remote get-url origin
	cmd := exec.Command("git", "-C", repoPath, "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("no git remote found: %w", err)
	}

	// Parse remote URL (handles SSH and HTTPS formats)
	return parseGitHubURL(strings.TrimSpace(string(output)))
}

// FetchIssues retrieves issues from the repository.
func FetchIssues(owner, repo string, cfg Config) ([]Issue, error) {
	if err := CheckCLI(); err != nil {
		return nil, err
	}

	// Build gh command with JSON output
	args := []string{
		"issue", "list",
		"--repo", fmt.Sprintf("%s/%s", owner, repo),
		"--state", string(cfg.DefaultState),
		"--limit", fmt.Sprintf("%d", cfg.MaxIssues),
		"--json", "number,title,state,url",
	}

	cmd := exec.Command("gh", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("failed to fetch issues: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to fetch issues: %w", err)
	}

	var issues []Issue
	if err := json.Unmarshal(output, &issues); err != nil {
		return nil, fmt.Errorf("failed to parse issues: %w", err)
	}

	return issues, nil
}

// FetchIssueBody retrieves the full body of a single issue.
func FetchIssueBody(owner, repo string, number int) (string, error) {
	if err := CheckCLI(); err != nil {
		return "", err
	}

	args := []string{
		"issue", "view",
		"--repo", fmt.Sprintf("%s/%s", owner, repo),
		fmt.Sprintf("%d", number),
		"--json", "body",
	}

	cmd := exec.Command("gh", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("failed to fetch issue body: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("failed to fetch issue body: %w", err)
	}

	var result struct {
		Body string `json:"body"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return "", fmt.Errorf("failed to parse issue body: %w", err)
	}

	return result.Body, nil
}

// parseGitHubURL extracts owner/repo from various GitHub URL formats.
func parseGitHubURL(url string) (owner, repo string, err error) {
	// Handle SSH format: git@github.com:owner/repo.git
	if strings.HasPrefix(url, "git@github.com:") {
		path := strings.TrimPrefix(url, "git@github.com:")
		path = strings.TrimSuffix(path, ".git")
		parts := strings.Split(path, "/")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid GitHub SSH URL: %s", url)
		}
		return parts[0], parts[1], nil
	}

	// Handle HTTPS format: https://github.com/owner/repo.git
	if strings.HasPrefix(url, "https://github.com/") {
		path := strings.TrimPrefix(url, "https://github.com/")
		path = strings.TrimSuffix(path, ".git")
		parts := strings.Split(path, "/")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid GitHub HTTPS URL: %s", url)
		}
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("not a GitHub repository: %s", url)
}
