// Package config handles loading and providing application configuration.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/christophergyman/claude-quick/internal/auth"
	"github.com/christophergyman/claude-quick/internal/constants"
	"github.com/christophergyman/claude-quick/internal/github"
	"github.com/christophergyman/claude-quick/internal/util"
	"gopkg.in/yaml.v3"
)

// ConfigSource represents where the config was loaded from
type ConfigSource int

const (
	ConfigSourceExecutable ConfigSource = iota // New location (co-located with executable)
	ConfigSourceLegacy                         // Legacy ~/.config location
	ConfigSourceDefault                        // No config file found, using defaults
)

// configInfo holds information about the resolved config location
var configInfo struct {
	Path   string
	Source ConfigSource
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
	AutoPushWorktree   *bool         `yaml:"auto_push_worktree,omitempty"`
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
		SearchPaths:        []string{util.HomeDir()},
		MaxDepth:           constants.DefaultMaxDepth,
		ExcludedDirs:       DefaultExcludedDirs(),
		DefaultSessionName: constants.DefaultSessionName,
		ContainerTimeout:   constants.DefaultContainerTimeout,
		GitHub:             github.DefaultConfig(),
	}
}

// executableDir returns the directory containing the resolved executable
// (follows symlinks to find the real location)
func executableDir() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	// Follow symlinks to get the real path
	realPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	return filepath.Dir(realPath), nil
}

// legacyConfigPath returns the legacy config path (~/.config/claude-quick/config.yaml)
func legacyConfigPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = filepath.Join(util.HomeDir(), ".config")
	}
	return filepath.Join(configDir, "claude-quick", "config.yaml")
}

// configPath resolves the config file path with the following priority:
// 1. Config file co-located with executable (new location)
// 2. Legacy ~/.config/claude-quick/config.yaml location
// Returns the path and the source type
func configPath() (string, ConfigSource) {
	// Try executable directory first (new location)
	if execDir, err := executableDir(); err == nil {
		newPath := filepath.Join(execDir, "claude-quick.yaml")
		if _, err := os.Stat(newPath); err == nil {
			return newPath, ConfigSourceExecutable
		}
	}

	// Fall back to legacy location
	legacyPath := legacyConfigPath()
	if _, err := os.Stat(legacyPath); err == nil {
		return legacyPath, ConfigSourceLegacy
	}

	// No config found - return new location path for error messages
	if execDir, err := executableDir(); err == nil {
		return filepath.Join(execDir, "claude-quick.yaml"), ConfigSourceDefault
	}

	// If executable dir cannot be determined, use legacy path as reference
	return legacyPath, ConfigSourceDefault
}

// Load reads the configuration from the config file
// Falls back to defaults if the file doesn't exist
// Prints deprecation warning if using legacy location
func Load() (*Config, error) {
	cfg := DefaultConfig()

	path, source := configPath()
	configInfo.Path = path
	configInfo.Source = source

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	// Print deprecation warning if using legacy location
	if source == ConfigSourceLegacy {
		targetPath := path
		if execDir, err := executableDir(); err == nil {
			targetPath = filepath.Join(execDir, "claude-quick.yaml")
		}
		fmt.Fprintf(os.Stderr, "Warning: Config loaded from deprecated location %s\n", path)
		fmt.Fprintf(os.Stderr, "         Consider moving to %s\n\n", targetPath)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Expand ~ in paths
	for i, p := range cfg.SearchPaths {
		cfg.SearchPaths[i] = util.ExpandPath(p)
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

// ConfigPath returns the path where the config file is/should be located
func ConfigPath() string {
	if configInfo.Path != "" {
		return configInfo.Path
	}
	path, _ := configPath()
	return path
}

// GetConfigSource returns the source of the loaded configuration
func GetConfigSource() ConfigSource {
	return configInfo.Source
}

// IsUsingLegacyConfig returns true if config was loaded from legacy location
func IsUsingLegacyConfig() bool {
	return configInfo.Source == ConfigSourceLegacy
}

// IsDarkMode returns the dark mode setting, defaulting to true if not set
func (c *Config) IsDarkMode() bool {
	if c.DarkMode == nil {
		return true // Default to dark mode for backwards compatibility
	}
	return *c.DarkMode
}

// IsAutoPushWorktree returns whether to auto-push new worktree branches upstream
func (c *Config) IsAutoPushWorktree() bool {
	if c.AutoPushWorktree == nil {
		return true // Default to enabled
	}
	return *c.AutoPushWorktree
}
