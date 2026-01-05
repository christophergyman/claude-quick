package auth

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveResult_HasErrors(t *testing.T) {
	tests := []struct {
		name     string
		errors   map[string]error
		expected bool
	}{
		{"no errors", map[string]error{}, false},
		{"nil errors", nil, false},
		{"one error", map[string]error{"key": os.ErrNotExist}, true},
		{"multiple errors", map[string]error{"key1": os.ErrNotExist, "key2": os.ErrPermission}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ResolveResult{Errors: tt.errors}
			if got := r.HasErrors(); got != tt.expected {
				t.Errorf("HasErrors() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestResolveResult_ErrorSummary(t *testing.T) {
	tests := []struct {
		name       string
		errors     map[string]error
		contains   []string
		isEmpty    bool
	}{
		{
			name:    "no errors - empty string",
			errors:  map[string]error{},
			isEmpty: true,
		},
		{
			name:     "single error",
			errors:   map[string]error{"API_KEY": os.ErrNotExist},
			contains: []string{"API_KEY", "not exist"},
		},
		{
			name:     "multiple errors",
			errors:   map[string]error{"KEY1": os.ErrNotExist, "KEY2": os.ErrPermission},
			contains: []string{"KEY1", "KEY2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ResolveResult{Errors: tt.errors}
			summary := r.ErrorSummary()

			if tt.isEmpty {
				if summary != "" {
					t.Errorf("ErrorSummary() = %q, want empty string", summary)
				}
				return
			}

			for _, substr := range tt.contains {
				if !strings.Contains(summary, substr) {
					t.Errorf("ErrorSummary() = %q, want to contain %q", summary, substr)
				}
			}
		})
	}
}

func TestConfig_ResolveLaunchCommand(t *testing.T) {
	tests := []struct {
		name          string
		config        *Config
		projectName   string
		globalDefault string
		expected      string
	}{
		{
			name:          "nil config returns global default",
			config:        nil,
			projectName:   "myproject",
			globalDefault: "claude",
			expected:      "claude",
		},
		{
			name:          "empty config returns global default",
			config:        &Config{},
			projectName:   "myproject",
			globalDefault: "claude",
			expected:      "claude",
		},
		{
			name: "project not found returns global default",
			config: &Config{
				Projects: map[string]ProjectAuth{
					"other": {LaunchCommand: "npm start"},
				},
			},
			projectName:   "myproject",
			globalDefault: "claude",
			expected:      "claude",
		},
		{
			name: "project found with launch command",
			config: &Config{
				Projects: map[string]ProjectAuth{
					"myproject": {LaunchCommand: "npm run dev"},
				},
			},
			projectName:   "myproject",
			globalDefault: "claude",
			expected:      "npm run dev",
		},
		{
			name: "project found without launch command returns global default",
			config: &Config{
				Projects: map[string]ProjectAuth{
					"myproject": {LaunchCommand: ""},
				},
			},
			projectName:   "myproject",
			globalDefault: "claude",
			expected:      "claude",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.ResolveLaunchCommand(tt.projectName, tt.globalDefault)
			if got != tt.expected {
				t.Errorf("ResolveLaunchCommand() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestConfig_Resolve_NilConfig(t *testing.T) {
	var config *Config
	result := config.Resolve("anyproject")

	if result == nil {
		t.Fatal("Resolve() returned nil")
	}
	if len(result.Credentials) != 0 {
		t.Errorf("Credentials = %v, want empty", result.Credentials)
	}
	if len(result.Errors) != 0 {
		t.Errorf("Errors = %v, want empty", result.Errors)
	}
}

func TestConfig_Resolve_EnvSource(t *testing.T) {
	// Set up test environment variable
	testEnvName := "TEST_CREDENTIAL_VAR"
	testEnvValue := "secret-value-123"
	os.Setenv(testEnvName, testEnvValue)
	defer os.Unsetenv(testEnvName)

	config := &Config{
		Credentials: []Credential{
			{Name: "MY_SECRET", Source: SourceEnv, Value: testEnvName},
		},
	}

	result := config.Resolve("anyproject")

	if result.HasErrors() {
		t.Errorf("Resolve() returned errors: %s", result.ErrorSummary())
	}
	if got := result.Credentials["MY_SECRET"]; got != testEnvValue {
		t.Errorf("Credentials[MY_SECRET] = %q, want %q", got, testEnvValue)
	}
}

func TestConfig_Resolve_EnvSource_NotSet(t *testing.T) {
	// Make sure the env var doesn't exist
	os.Unsetenv("NONEXISTENT_TEST_VAR")

	config := &Config{
		Credentials: []Credential{
			{Name: "MY_SECRET", Source: SourceEnv, Value: "NONEXISTENT_TEST_VAR"},
		},
	}

	result := config.Resolve("anyproject")

	if !result.HasErrors() {
		t.Error("Resolve() should have returned errors for unset env var")
	}
	if _, exists := result.Errors["MY_SECRET"]; !exists {
		t.Error("Expected error for MY_SECRET")
	}
}

func TestConfig_Resolve_FileSource(t *testing.T) {
	// Create a temporary file with credential content
	tmpFile, err := os.CreateTemp("", "test-credential-*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	testValue := "file-secret-value"
	if _, err := tmpFile.WriteString(testValue + "\n"); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	config := &Config{
		Credentials: []Credential{
			{Name: "FILE_SECRET", Source: SourceFile, Value: tmpFile.Name()},
		},
	}

	result := config.Resolve("anyproject")

	if result.HasErrors() {
		t.Errorf("Resolve() returned errors: %s", result.ErrorSummary())
	}
	// Note: The value should be trimmed
	if got := result.Credentials["FILE_SECRET"]; got != testValue {
		t.Errorf("Credentials[FILE_SECRET] = %q, want %q", got, testValue)
	}
}

func TestConfig_Resolve_FileSource_NotFound(t *testing.T) {
	config := &Config{
		Credentials: []Credential{
			{Name: "MISSING_FILE", Source: SourceFile, Value: "/nonexistent/path/to/credential"},
		},
	}

	result := config.Resolve("anyproject")

	if !result.HasErrors() {
		t.Error("Resolve() should have returned errors for missing file")
	}
	if _, exists := result.Errors["MISSING_FILE"]; !exists {
		t.Error("Expected error for MISSING_FILE")
	}
}

func TestConfig_Resolve_FileSource_WithTilde(t *testing.T) {
	// Create a temp file in a known location
	tmpDir, err := os.MkdirTemp("", "test-tilde-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, "credential")
	testValue := "tilde-secret"
	if err := os.WriteFile(tmpFile, []byte(testValue), 0600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	// Test with absolute path (tilde expansion happens in resolveFileSource)
	config := &Config{
		Credentials: []Credential{
			{Name: "TILDE_SECRET", Source: SourceFile, Value: tmpFile},
		},
	}

	result := config.Resolve("anyproject")

	if result.HasErrors() {
		t.Errorf("Resolve() returned errors: %s", result.ErrorSummary())
	}
	if got := result.Credentials["TILDE_SECRET"]; got != testValue {
		t.Errorf("Credentials[TILDE_SECRET] = %q, want %q", got, testValue)
	}
}

func TestConfig_Resolve_ProjectOverride(t *testing.T) {
	// Set up environment variables
	os.Setenv("GLOBAL_VAR", "global-value")
	os.Setenv("PROJECT_VAR", "project-value")
	defer os.Unsetenv("GLOBAL_VAR")
	defer os.Unsetenv("PROJECT_VAR")

	config := &Config{
		Credentials: []Credential{
			{Name: "SECRET", Source: SourceEnv, Value: "GLOBAL_VAR"},
		},
		Projects: map[string]ProjectAuth{
			"special-project": {
				Credentials: []Credential{
					{Name: "SECRET", Source: SourceEnv, Value: "PROJECT_VAR"},
				},
			},
		},
	}

	// Test with non-matching project (should use global)
	result := config.Resolve("other-project")
	if got := result.Credentials["SECRET"]; got != "global-value" {
		t.Errorf("Resolve(other-project) SECRET = %q, want %q", got, "global-value")
	}

	// Test with matching project (should use project override)
	result = config.Resolve("special-project")
	if got := result.Credentials["SECRET"]; got != "project-value" {
		t.Errorf("Resolve(special-project) SECRET = %q, want %q", got, "project-value")
	}
}

func TestConfig_Resolve_MultipleCredentials(t *testing.T) {
	os.Setenv("TEST_VAR1", "value1")
	os.Setenv("TEST_VAR2", "value2")
	defer os.Unsetenv("TEST_VAR1")
	defer os.Unsetenv("TEST_VAR2")

	config := &Config{
		Credentials: []Credential{
			{Name: "CRED1", Source: SourceEnv, Value: "TEST_VAR1"},
			{Name: "CRED2", Source: SourceEnv, Value: "TEST_VAR2"},
		},
	}

	result := config.Resolve("anyproject")

	if result.HasErrors() {
		t.Errorf("Resolve() returned errors: %s", result.ErrorSummary())
	}
	if len(result.Credentials) != 2 {
		t.Errorf("Credentials count = %d, want 2", len(result.Credentials))
	}
	if got := result.Credentials["CRED1"]; got != "value1" {
		t.Errorf("Credentials[CRED1] = %q, want %q", got, "value1")
	}
	if got := result.Credentials["CRED2"]; got != "value2" {
		t.Errorf("Credentials[CRED2] = %q, want %q", got, "value2")
	}
}

func TestConfig_Resolve_PartialFailure(t *testing.T) {
	// One valid, one invalid credential
	os.Setenv("VALID_VAR", "valid-value")
	os.Unsetenv("INVALID_VAR")
	defer os.Unsetenv("VALID_VAR")

	config := &Config{
		Credentials: []Credential{
			{Name: "VALID", Source: SourceEnv, Value: "VALID_VAR"},
			{Name: "INVALID", Source: SourceEnv, Value: "INVALID_VAR"},
		},
	}

	result := config.Resolve("anyproject")

	// Should have one success and one error
	if got := result.Credentials["VALID"]; got != "valid-value" {
		t.Errorf("Credentials[VALID] = %q, want %q", got, "valid-value")
	}
	if _, exists := result.Credentials["INVALID"]; exists {
		t.Error("INVALID should not be in Credentials")
	}
	if !result.HasErrors() {
		t.Error("Should have errors for INVALID credential")
	}
	if _, exists := result.Errors["INVALID"]; !exists {
		t.Error("Expected error for INVALID")
	}
}
