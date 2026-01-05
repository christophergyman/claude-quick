package auth

import (
	"strings"
	"testing"
)

func TestCredential_Validate(t *testing.T) {
	tests := []struct {
		name       string
		credential Credential
		wantErr    bool
		errContain string
	}{
		{
			name: "valid file credential",
			credential: Credential{
				Name:   "API_KEY",
				Source: SourceFile,
				Value:  "~/.secrets/api_key",
			},
			wantErr: false,
		},
		{
			name: "valid env credential",
			credential: Credential{
				Name:   "GITHUB_TOKEN",
				Source: SourceEnv,
				Value:  "GH_TOKEN",
			},
			wantErr: false,
		},
		{
			name: "valid command credential",
			credential: Credential{
				Name:   "SECRET",
				Source: SourceCommand,
				Value:  "op read op://vault/item/password",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			credential: Credential{
				Name:   "",
				Source: SourceFile,
				Value:  "/path/to/file",
			},
			wantErr:    true,
			errContain: "name is required",
		},
		{
			name: "empty source",
			credential: Credential{
				Name:   "API_KEY",
				Source: "",
				Value:  "/path/to/file",
			},
			wantErr:    true,
			errContain: "source is required",
		},
		{
			name: "invalid source type",
			credential: Credential{
				Name:   "API_KEY",
				Source: "invalid",
				Value:  "/path/to/file",
			},
			wantErr:    true,
			errContain: "invalid source type",
		},
		{
			name: "empty value",
			credential: Credential{
				Name:   "API_KEY",
				Source: SourceFile,
				Value:  "",
			},
			wantErr:    true,
			errContain: "value is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.credential.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Credential.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
				t.Errorf("Credential.Validate() error = %v, want error containing %q", err, tt.errContain)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name       string
		config     Config
		wantErr    bool
		errContain string
	}{
		{
			name:    "empty config is valid",
			config:  Config{},
			wantErr: false,
		},
		{
			name: "valid global credentials",
			config: Config{
				Credentials: []Credential{
					{Name: "API_KEY", Source: SourceFile, Value: "~/.secrets/key"},
					{Name: "TOKEN", Source: SourceEnv, Value: "MY_TOKEN"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid project credentials",
			config: Config{
				Projects: map[string]ProjectAuth{
					"my-project": {
						Credentials: []Credential{
							{Name: "API_KEY", Source: SourceFile, Value: "~/.secrets/key"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid global credential",
			config: Config{
				Credentials: []Credential{
					{Name: "", Source: SourceFile, Value: "~/.secrets/key"},
				},
			},
			wantErr:    true,
			errContain: "auth.credentials[0]",
		},
		{
			name: "invalid project credential",
			config: Config{
				Projects: map[string]ProjectAuth{
					"my-project": {
						Credentials: []Credential{
							{Name: "API_KEY", Source: "invalid", Value: "~/.secrets/key"},
						},
					},
				},
			},
			wantErr:    true,
			errContain: "auth.projects.my-project.credentials[0]",
		},
		{
			name: "mixed valid and invalid credentials",
			config: Config{
				Credentials: []Credential{
					{Name: "API_KEY", Source: SourceFile, Value: "~/.secrets/key"},
					{Name: "TOKEN", Source: SourceEnv, Value: ""},
				},
			},
			wantErr:    true,
			errContain: "auth.credentials[1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
				t.Errorf("Config.Validate() error = %v, want error containing %q", err, tt.errContain)
			}
		})
	}
}

func TestSourceTypes(t *testing.T) {
	// Verify all source types are in ValidSourceTypes
	sourceTypes := []SourceType{SourceFile, SourceEnv, SourceCommand}
	for _, st := range sourceTypes {
		if !ValidSourceTypes[st] {
			t.Errorf("SourceType %q not in ValidSourceTypes", st)
		}
	}

	// Verify ValidSourceTypes has exactly 3 entries
	if len(ValidSourceTypes) != 3 {
		t.Errorf("ValidSourceTypes has %d entries, expected 3", len(ValidSourceTypes))
	}
}

func TestSourceTypeConstants(t *testing.T) {
	// Verify source type string values match expected
	tests := []struct {
		name     string
		source   SourceType
		expected string
	}{
		{"file source", SourceFile, "file"},
		{"env source", SourceEnv, "env"},
		{"command source", SourceCommand, "command"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.source) != tt.expected {
				t.Errorf("SourceType = %q, want %q", tt.source, tt.expected)
			}
		})
	}
}
