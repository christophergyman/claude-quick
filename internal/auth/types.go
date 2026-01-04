// Package auth handles authentication credential management and injection.
package auth

import "fmt"

// SourceType defines how to retrieve a credential value.
type SourceType string

const (
	// SourceFile reads the credential from a file.
	SourceFile SourceType = "file"
	// SourceEnv reads the credential from an environment variable.
	SourceEnv SourceType = "env"
	// SourceCommand runs a command and uses its output as the credential.
	SourceCommand SourceType = "command"
)

// ValidSourceTypes contains all valid source type values.
var ValidSourceTypes = map[SourceType]bool{
	SourceFile:    true,
	SourceEnv:     true,
	SourceCommand: true,
}

// Credential defines a single authentication credential.
type Credential struct {
	// Name is the environment variable name to expose in the container.
	Name string `yaml:"name"`
	// Source defines how to retrieve the credential value.
	Source SourceType `yaml:"source"`
	// Value is the path (for file), env var name (for env), or command (for command).
	Value string `yaml:"value"`
}

// ProjectAuth defines project-specific authentication overrides.
type ProjectAuth struct {
	// Credentials overrides the global credentials for this project.
	Credentials []Credential `yaml:"credentials,omitempty"`
}

// Config holds the authentication configuration.
type Config struct {
	// Credentials defines the global credentials to inject into all containers.
	Credentials []Credential `yaml:"credentials,omitempty"`
	// Projects defines project-specific credential overrides.
	Projects map[string]ProjectAuth `yaml:"projects,omitempty"`
}

// Validate checks that the auth configuration is valid.
func (c *Config) Validate() error {
	// Validate global credentials
	for i, cred := range c.Credentials {
		if err := cred.Validate(); err != nil {
			return fmt.Errorf("auth.credentials[%d]: %w", i, err)
		}
	}

	// Validate project-specific credentials
	for projName, proj := range c.Projects {
		for i, cred := range proj.Credentials {
			if err := cred.Validate(); err != nil {
				return fmt.Errorf("auth.projects.%s.credentials[%d]: %w", projName, i, err)
			}
		}
	}

	return nil
}

// Validate checks that the credential configuration is valid.
func (c *Credential) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}

	if c.Source == "" {
		return fmt.Errorf("source is required")
	}

	if !ValidSourceTypes[c.Source] {
		return fmt.Errorf("invalid source type %q (must be file, env, or command)", c.Source)
	}

	if c.Value == "" {
		return fmt.Errorf("value is required")
	}

	return nil
}
