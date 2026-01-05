package util

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("could not determine home directory: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "no tilde - absolute path",
			input:    "/usr/local/bin",
			expected: "/usr/local/bin",
		},
		{
			name:     "no tilde - relative path",
			input:    "some/relative/path",
			expected: "some/relative/path",
		},
		{
			name:     "tilde only",
			input:    "~",
			expected: homeDir,
		},
		{
			name:     "tilde with slash",
			input:    "~/",
			expected: homeDir,
		},
		{
			name:     "tilde with path",
			input:    "~/documents",
			expected: filepath.Join(homeDir, "documents"),
		},
		{
			name:     "tilde with nested path",
			input:    "~/.config/claude-quick/config.yaml",
			expected: filepath.Join(homeDir, ".config/claude-quick/config.yaml"),
		},
		{
			name:     "tilde in middle - not expanded",
			input:    "/path/to/~user/file",
			expected: "/path/to/~user/file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandPath(tt.input)
			if result != tt.expected {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExpandPath_TildeExpansion(t *testing.T) {
	result := ExpandPath("~/test")

	// Should start with either home dir or temp dir (fallback)
	homeDir, _ := os.UserHomeDir()
	tempDir := os.TempDir()

	if !strings.HasPrefix(result, homeDir) && !strings.HasPrefix(result, tempDir) {
		t.Errorf("ExpandPath(\"~/test\") = %q, expected to start with home dir %q or temp dir %q",
			result, homeDir, tempDir)
	}

	// Should end with /test
	if !strings.HasSuffix(result, "test") {
		t.Errorf("ExpandPath(\"~/test\") = %q, expected to end with \"test\"", result)
	}
}

func TestGetHomeDir(t *testing.T) {
	result := getHomeDir()

	// Should return a non-empty string
	if result == "" {
		t.Error("getHomeDir() returned empty string")
	}

	// Should be a valid directory path (either home or temp)
	homeDir, homeErr := os.UserHomeDir()
	tempDir := os.TempDir()

	if homeErr == nil && result != homeDir {
		// If we can get home dir, result should match
		t.Errorf("getHomeDir() = %q, want %q", result, homeDir)
	} else if homeErr != nil && result != tempDir {
		// If home dir fails, should fall back to temp
		t.Errorf("getHomeDir() = %q, want temp dir %q", result, tempDir)
	}
}
