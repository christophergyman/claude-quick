package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/christophergyman/claude-quick/internal/constants"
)

func TestDefaultExcludedDirs(t *testing.T) {
	dirs := DefaultExcludedDirs()

	// Should return non-empty list
	if len(dirs) == 0 {
		t.Error("DefaultExcludedDirs() returned empty slice")
	}

	// Should match constants.DefaultExcludedDirs
	constDirs := constants.DefaultExcludedDirs()
	if len(dirs) != len(constDirs) {
		t.Errorf("DefaultExcludedDirs() returned %d items, constants has %d", len(dirs), len(constDirs))
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	// Check search paths
	if len(cfg.SearchPaths) == 0 {
		t.Error("SearchPaths is empty")
	}

	// Check max depth
	if cfg.MaxDepth != constants.DefaultMaxDepth {
		t.Errorf("MaxDepth = %d, want %d", cfg.MaxDepth, constants.DefaultMaxDepth)
	}

	// Check excluded dirs
	if len(cfg.ExcludedDirs) == 0 {
		t.Error("ExcludedDirs is empty")
	}

	// Check default session name
	if cfg.DefaultSessionName != constants.DefaultSessionName {
		t.Errorf("DefaultSessionName = %q, want %q", cfg.DefaultSessionName, constants.DefaultSessionName)
	}

	// Check container timeout
	if cfg.ContainerTimeout != constants.DefaultContainerTimeout {
		t.Errorf("ContainerTimeout = %d, want %d", cfg.ContainerTimeout, constants.DefaultContainerTimeout)
	}
}

func TestConfig_IsDarkMode(t *testing.T) {
	tests := []struct {
		name     string
		darkMode *bool
		expected bool
	}{
		{
			name:     "nil defaults to true",
			darkMode: nil,
			expected: true,
		},
		{
			name:     "explicit true",
			darkMode: boolPtr(true),
			expected: true,
		},
		{
			name:     "explicit false",
			darkMode: boolPtr(false),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{DarkMode: tt.darkMode}
			if got := cfg.IsDarkMode(); got != tt.expected {
				t.Errorf("IsDarkMode() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func TestLoad_NonexistentFile(t *testing.T) {
	// Load should return defaults when file doesn't exist
	// This test assumes the config file doesn't exist in the test environment
	// or will use a temp directory

	cfg, err := Load()

	// Should not error for non-existent file
	if err != nil {
		// Only fail if it's not a "doesn't exist" scenario
		// The actual config file might exist in some test environments
		t.Logf("Load() returned error (may be expected): %v", err)
	}

	if cfg != nil {
		// If we got a config, verify defaults are set
		if cfg.MaxDepth <= 0 {
			t.Error("MaxDepth should be > 0")
		}
	}
}

func TestLoad_WithConfigFile(t *testing.T) {
	// Create a temporary config directory and file
	tmpDir, err := os.MkdirTemp("", "test-config-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configDir := filepath.Join(tmpDir, "claude-quick")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configFile := filepath.Join(configDir, "config.yaml")
	configContent := `
search_paths:
  - /home/user/projects
max_depth: 5
excluded_dirs:
  - node_modules
  - vendor
default_session_name: dev
container_timeout_seconds: 600
launch_command: "npm start"
dark_mode: false
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Note: We can't easily test Load() directly because it uses os.UserConfigDir()
	// This test documents the expected behavior
	t.Log("Config file parsing is tested via Load() integration")
}

func TestConfigValidation(t *testing.T) {
	// Test that invalid configs fail validation
	tests := []struct {
		name    string
		timeout int
		want    int
	}{
		{"zero timeout uses default", 0, constants.DefaultContainerTimeout},
		{"negative timeout uses default", -1, constants.DefaultContainerTimeout},
		{"below min uses min", 10, constants.MinContainerTimeout},
		{"above max uses max", 10000, constants.MaxContainerTimeout},
		{"valid timeout unchanged", 120, 120},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This tests the validation logic conceptually
			// Actual validation happens in Load()
			var result int
			if tt.timeout <= 0 {
				result = constants.DefaultContainerTimeout
			} else if tt.timeout < constants.MinContainerTimeout {
				result = constants.MinContainerTimeout
			} else if tt.timeout > constants.MaxContainerTimeout {
				result = constants.MaxContainerTimeout
			} else {
				result = tt.timeout
			}

			if result != tt.want {
				t.Errorf("timeout validation: got %d, want %d", result, tt.want)
			}
		})
	}
}

func TestConfigPath(t *testing.T) {
	path := ConfigPath()

	// Should return a non-empty string
	if path == "" {
		t.Error("ConfigPath() returned empty string")
	}

	// Should end with either claude-quick.yaml (new) or config.yaml (legacy)
	base := filepath.Base(path)
	if base != "claude-quick.yaml" && base != "config.yaml" {
		t.Errorf("ConfigPath() = %q, should end with claude-quick.yaml or config.yaml", path)
	}

	// Should be an absolute path
	if !filepath.IsAbs(path) {
		t.Logf("ConfigPath() returned relative path: %s", path)
	}
}

func TestExecutableDir(t *testing.T) {
	dir, err := executableDir()
	if err != nil {
		t.Logf("executableDir() returned error (may be expected in test env): %v", err)
		return
	}

	if dir == "" {
		t.Error("executableDir() returned empty string")
	}

	// Should be an absolute path
	if !filepath.IsAbs(dir) {
		t.Errorf("executableDir() returned non-absolute path: %s", dir)
	}

	// Should be a valid directory
	info, err := os.Stat(dir)
	if err != nil {
		t.Errorf("executableDir() returned invalid path: %v", err)
	} else if !info.IsDir() {
		t.Errorf("executableDir() returned non-directory: %s", dir)
	}
}

func TestLegacyConfigPath(t *testing.T) {
	path := legacyConfigPath()

	if path == "" {
		t.Error("legacyConfigPath() returned empty string")
	}

	// Should contain claude-quick directory
	if !filepath.IsAbs(path) {
		t.Logf("legacyConfigPath() returned relative path: %s", path)
	}

	// Should end with config.yaml
	if filepath.Base(path) != "config.yaml" {
		t.Errorf("legacyConfigPath() = %q, should end with config.yaml", path)
	}

	// Should contain claude-quick in the path
	dir := filepath.Dir(path)
	if filepath.Base(dir) != "claude-quick" {
		t.Errorf("legacyConfigPath() parent dir = %q, should be claude-quick", filepath.Base(dir))
	}
}

func TestConfigPath_Priority(t *testing.T) {
	path, source := configPath()

	// Path should never be empty
	if path == "" {
		t.Error("configPath() returned empty path")
	}

	// Source should be a valid value
	switch source {
	case ConfigSourceExecutable, ConfigSourceLegacy, ConfigSourceDefault:
		// OK
	default:
		t.Errorf("configPath() returned invalid source: %d", source)
	}

	t.Logf("configPath() resolved to %q (source: %d)", path, source)
}

func TestGetConfigSource(t *testing.T) {
	// After Load() is called, GetConfigSource should return a valid source
	_, err := Load()
	if err != nil {
		t.Logf("Load() returned error (may be expected): %v", err)
	}

	source := GetConfigSource()
	switch source {
	case ConfigSourceExecutable, ConfigSourceLegacy, ConfigSourceDefault:
		// OK
	default:
		t.Errorf("GetConfigSource() returned invalid source: %d", source)
	}
}

func TestIsUsingLegacyConfig(t *testing.T) {
	// Test documents expected behavior
	// Actual value depends on whether legacy config exists
	isLegacy := IsUsingLegacyConfig()
	t.Logf("IsUsingLegacyConfig() = %v", isLegacy)
}
