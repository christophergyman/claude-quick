// Package util provides shared utility functions used across packages.
package util

import (
	"os"
	"path/filepath"
)

// ExpandPath expands ~ to the user's home directory.
// If the home directory cannot be determined, it falls back to the system temp directory.
// This ensures the returned path is always usable.
func ExpandPath(path string) string {
	if len(path) == 0 {
		return path
	}
	if path[0] == '~' {
		return filepath.Join(HomeDir(), path[1:])
	}
	return path
}

// HomeDir returns the user's home directory with fallback to temp directory.
func HomeDir() string {
	if homeDir, err := os.UserHomeDir(); err == nil {
		return homeDir
	}
	return os.TempDir()
}
