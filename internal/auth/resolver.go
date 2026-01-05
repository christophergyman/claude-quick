package auth

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/christophergyman/claude-quick/internal/util"
)

// ResolveResult contains resolved credentials and any errors encountered.
type ResolveResult struct {
	// Credentials maps environment variable names to their resolved values.
	Credentials map[string]string
	// Errors contains resolution errors keyed by credential name.
	Errors map[string]error
}

// HasErrors returns true if any credentials failed to resolve.
func (r *ResolveResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// ErrorSummary returns a human-readable summary of resolution errors.
func (r *ResolveResult) ErrorSummary() string {
	if !r.HasErrors() {
		return ""
	}

	var parts []string
	for name, err := range r.Errors {
		parts = append(parts, fmt.Sprintf("%s: %v", name, err))
	}
	return strings.Join(parts, "; ")
}

// Resolve resolves all credentials for a given project.
// If the project has specific overrides, those are used instead of global credentials.
// Returns a ResolveResult containing resolved credentials and any errors.
func (c *Config) Resolve(projectName string) *ResolveResult {
	result := &ResolveResult{
		Credentials: make(map[string]string),
		Errors:      make(map[string]error),
	}

	if c == nil {
		return result
	}

	// Determine which credentials to use
	creds := c.Credentials
	if proj, ok := c.Projects[projectName]; ok && len(proj.Credentials) > 0 {
		creds = proj.Credentials
	}

	for _, cred := range creds {
		value, err := resolveCredential(cred)
		if err != nil {
			result.Errors[cred.Name] = err
			continue
		}
		if value != "" {
			result.Credentials[cred.Name] = value
		}
	}

	return result
}

// ResolveLaunchCommand returns the launch command for a project.
// Returns the project-specific command if set, otherwise the global default.
func (c *Config) ResolveLaunchCommand(projectName, globalDefault string) string {
	if c != nil {
		if proj, ok := c.Projects[projectName]; ok && proj.LaunchCommand != "" {
			return proj.LaunchCommand
		}
	}
	return globalDefault
}

// resolveCredential resolves a single credential from its source.
func resolveCredential(cred Credential) (string, error) {
	switch cred.Source {
	case SourceFile:
		return resolveFileSource(cred.Value)
	case SourceEnv:
		return resolveEnvSource(cred.Value)
	case SourceCommand:
		return resolveCommandSource(cred.Value)
	default:
		return "", fmt.Errorf("unknown source type: %s", cred.Source)
	}
}

// resolveFileSource reads a credential from a file.
func resolveFileSource(path string) (string, error) {
	expandedPath := util.ExpandPath(path)
	data, err := os.ReadFile(expandedPath)
	if err != nil {
		return "", fmt.Errorf("failed to read credential file %s: %w", path, err)
	}
	return strings.TrimSpace(string(data)), nil
}

// resolveEnvSource reads a credential from an environment variable.
// Returns an error if the environment variable is not set.
func resolveEnvSource(name string) (string, error) {
	value, exists := os.LookupEnv(name)
	if !exists {
		return "", fmt.Errorf("environment variable %s is not set", name)
	}
	return value, nil
}

// resolveCommandSource runs a command and returns its output as the credential.
func resolveCommandSource(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run credential command: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}
