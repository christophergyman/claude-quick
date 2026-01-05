// Package config handles loading and providing application configuration.
package config

import (
	"os"
	"path/filepath"

	"github.com/christophergyman/claude-quick/internal/auth"
	"github.com/christophergyman/claude-quick/internal/constants"
	"github.com/christophergyman/claude-quick/internal/github"
	"gopkg.in/yaml.v3"
)

// getHomeDir returns the user's home directory with fallback to /tmp
func getHomeDir() string {
	if homeDir, err := os.UserHomeDir(); err == nil {
		return homeDir
	}
	// Fallback to /tmp if home directory cannot be determined
	return os.TempDir()
}

// Config holds the application configuration
type Config struct {
	SearchPaths        []string      `yaml:"search_paths"`
	MaxDepth           int           `yaml:"max_depth"`
	ExcludedDirs       []string      `yaml:"excluded_dirs"`
	DefaultSessionName string        `yaml:"default_session_name"`
	ContainerTimeout   int           `yaml:"container_timeout_seconds"`
	LaunchCommand      string        `yaml:"launch_command,omitempty"`
	DarkMode           *bool         `yaml:"dark_mode,omitempty"`
	Auth               auth.Config   `yaml:"auth,omitempty"`
	GitHub             github.Config `yaml:"github,omitempty"`
}

// DefaultExcludedDirs returns the default directories to exclude from scanning
func DefaultExcludedDirs() []string {
	return constants.DefaultExcludedDirs()
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		SearchPaths:        []string{getHomeDir()},
		MaxDepth:           constants.DefaultMaxDepth,
		ExcludedDirs:       DefaultExcludedDirs(),
		DefaultSessionName: constants.DefaultSessionName,
		ContainerTimeout:   constants.DefaultContainerTimeout,
		GitHub:             github.DefaultConfig(),
	}
}

// configPath returns the path to the config file
func configPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = filepath.Join(getHomeDir(), ".config")
	}
	return filepath.Join(configDir, "claude-quick", "config.yaml")
}

// Load reads the configuration from the config file
// Falls back to defaults if the file doesn't exist
func Load() (*Config, error) {
	cfg := DefaultConfig()

	path := configPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Expand ~ in paths
	for i, p := range cfg.SearchPaths {
		cfg.SearchPaths[i] = expandPath(p)
	}

	// Ensure reasonable defaults
	if cfg.MaxDepth <= 0 {
		cfg.MaxDepth = constants.DefaultMaxDepth
	}

	// Use default excluded dirs if none specified
	if len(cfg.ExcludedDirs) == 0 {
		cfg.ExcludedDirs = DefaultExcludedDirs()
	}

	// Ensure default session name
	if cfg.DefaultSessionName == "" {
		cfg.DefaultSessionName = constants.DefaultSessionName
	}

	// Ensure reasonable timeout (minimum 30 seconds, max 30 minutes)
	if cfg.ContainerTimeout <= 0 {
		cfg.ContainerTimeout = constants.DefaultContainerTimeout
	} else if cfg.ContainerTimeout < constants.MinContainerTimeout {
		cfg.ContainerTimeout = constants.MinContainerTimeout
	} else if cfg.ContainerTimeout > constants.MaxContainerTimeout {
		cfg.ContainerTimeout = constants.MaxContainerTimeout
	}

	// Validate auth configuration
	if err := cfg.Auth.Validate(); err != nil {
		return nil, err
	}

	// Ensure GitHub config has sensible defaults
	if cfg.GitHub.MaxIssues <= 0 {
		cfg.GitHub.MaxIssues = constants.DefaultMaxIssues
	}
	if cfg.GitHub.BranchPrefix == "" {
		cfg.GitHub.BranchPrefix = constants.DefaultBranchPrefix
	}
	if cfg.GitHub.DefaultState == "" {
		cfg.GitHub.DefaultState = github.IssueStateOpen
	}

	return cfg, nil
}

// expandPath expands ~ to the user's home directory
func expandPath(path string) string {
	if len(path) == 0 {
		return path
	}
	if path[0] == '~' {
		return filepath.Join(getHomeDir(), path[1:])
	}
	return path
}

// ConfigPath returns the path where the config file should be located
func ConfigPath() string {
	return configPath()
}

// IsDarkMode returns the dark mode setting, defaulting to true if not set
func (c *Config) IsDarkMode() bool {
	if c.DarkMode == nil {
		return true // Default to dark mode for backwards compatibility
	}
	return *c.DarkMode
}
