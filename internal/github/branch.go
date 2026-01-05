package github

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/christophergyman/claude-quick/internal/constants"
)

// GenerateBranchName creates a branch name from an issue.
// Format: <prefix><number>-<slugified-title>
// Example: issue-42-fix-login-bug
func GenerateBranchName(issue *Issue, prefix string) string {
	slug := slugifyTitle(issue.Title)
	branchName := fmt.Sprintf("%s%d-%s", prefix, issue.Number, slug)

	// Truncate if too long
	if len(branchName) > constants.MaxBranchNameLength {
		branchName = branchName[:constants.MaxBranchNameLength]
		// Ensure we don't end with a hyphen
		branchName = strings.TrimSuffix(branchName, "-")
	}

	return branchName
}

// slugifyTitle converts a title to a URL-safe slug.
func slugifyTitle(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Replace spaces and underscores with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")

	// Remove non-alphanumeric characters (except hyphen)
	var result strings.Builder
	for _, r := range slug {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			result.WriteRune(r)
		}
	}
	slug = result.String()

	// Replace multiple hyphens with single hyphen
	re := regexp.MustCompile(`-+`)
	slug = re.ReplaceAllString(slug, "-")

	// Trim leading/trailing hyphens
	slug = strings.Trim(slug, "-")

	return slug
}
