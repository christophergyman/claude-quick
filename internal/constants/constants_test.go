package constants

import (
	"testing"
)

func TestIsReservedBranchName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"main is reserved", "main", true},
		{"master is reserved", "master", true},
		{"feature branch not reserved", "feature/auth", false},
		{"develop not reserved", "develop", false},
		{"empty string not reserved", "", false},
		{"Main (uppercase) not reserved", "Main", false},
		{"MASTER (uppercase) not reserved", "MASTER", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsReservedBranchName(tt.input)
			if result != tt.expected {
				t.Errorf("IsReservedBranchName(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDefaultExcludedDirs(t *testing.T) {
	dirs := DefaultExcludedDirs()

	// Should return non-empty list
	if len(dirs) == 0 {
		t.Error("DefaultExcludedDirs() returned empty slice")
	}

	// Check for expected directories
	expected := map[string]bool{
		"node_modules": true,
		"vendor":       true,
		".git":         true,
		"__pycache__":  true,
		"venv":         true,
		".venv":        true,
		"dist":         true,
		"build":        true,
		".cache":       true,
	}

	for _, dir := range dirs {
		if !expected[dir] {
			t.Errorf("DefaultExcludedDirs() contains unexpected dir %q", dir)
		}
		delete(expected, dir)
	}

	// Check we got all expected directories
	for dir := range expected {
		t.Errorf("DefaultExcludedDirs() missing expected dir %q", dir)
	}
}

func TestReservedBranchNames(t *testing.T) {
	// Verify the list contains expected values
	if len(ReservedBranchNames) != 2 {
		t.Errorf("ReservedBranchNames has %d elements, expected 2", len(ReservedBranchNames))
	}

	expectedNames := map[string]bool{"main": true, "master": true}
	for _, name := range ReservedBranchNames {
		if !expectedNames[name] {
			t.Errorf("ReservedBranchNames contains unexpected %q", name)
		}
	}
}

func TestConstants(t *testing.T) {
	// Verify constants have sensible values
	tests := []struct {
		name     string
		value    int
		minValue int
		maxValue int
	}{
		{"DefaultContainerTimeout", DefaultContainerTimeout, 1, 3600},
		{"MinContainerTimeout", MinContainerTimeout, 1, DefaultContainerTimeout},
		{"MaxContainerTimeout", MaxContainerTimeout, DefaultContainerTimeout, 7200},
		{"DefaultMaxDepth", DefaultMaxDepth, 1, 10},
		{"TextInputCharLimit", TextInputCharLimit, 1, 200},
		{"TextInputWidth", TextInputWidth, 1, 100},
		{"SHATruncateLength", SHATruncateLength, 4, 40},
		{"DefaultPathTruncateLen", DefaultPathTruncateLen, 10, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value < tt.minValue || tt.value > tt.maxValue {
				t.Errorf("%s = %d, expected between %d and %d", tt.name, tt.value, tt.minValue, tt.maxValue)
			}
		})
	}
}

func TestStringConstants(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"DevcontainerDir", DevcontainerDir},
		{"DevcontainerConfigFile", DevcontainerConfigFile},
		{"DefaultSessionName", DefaultSessionName},
		{"DefaultWorktreePlaceholder", DefaultWorktreePlaceholder},
		{"DefaultBranchUnknown", DefaultBranchUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value == "" {
				t.Errorf("%s is empty", tt.name)
			}
		})
	}
}
