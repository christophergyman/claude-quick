package devcontainer

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/christophergyman/claude-quick/internal/auth"
)

// ListTmuxSessions lists tmux sessions inside the container
// Returns empty slice (not nil) if no sessions exist
func ListTmuxSessions(projectPath string) ([]string, error) {
	output, err := execInContainer(projectPath, "tmux", "list-sessions", "-F", "#{session_name}:#{session_attached}")
	if err != nil {
		// Exit code 1 means no sessions - return empty slice, not error
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return []string{}, nil
		}
		return nil, fmt.Errorf("ListTmuxSessions: %w", err)
	}

	var sessions []string
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line != "" {
			sessions = append(sessions, line)
		}
	}
	return sessions, nil
}

// CreateTmuxSession creates a new tmux session in the container
func CreateTmuxSession(projectPath, sessionName string) error {
	if err := execInContainerWithStderr(projectPath, "failed to create tmux session",
		"tmux", "new-session", "-d", "-s", sessionName); err != nil {
		return err
	}

	// Inject auth credentials into tmux session environment
	// Using setenv makes env vars available to all windows/panes in the session
	injectTmuxSessionEnv(projectPath, sessionName)

	return nil
}

// injectTmuxSessionEnv reads credentials from the auth file and sets them as tmux session env vars.
// Uses "tmux setenv" which propagates to all new windows/panes in the session.
func injectTmuxSessionEnv(projectPath, sessionName string) {
	creds := readCredentialFile(projectPath)
	for name, value := range creds {
		// tmux setenv -t session NAME value
		execInContainer(projectPath, "tmux", "setenv", "-t", sessionName, name, value)
	}
}

// readCredentialFile parses the .claude-quick-auth file and returns env var name/value pairs.
func readCredentialFile(projectPath string) map[string]string {
	result := make(map[string]string)

	filePath := filepath.Join(projectPath, auth.CredFileName)
	file, err := os.Open(filePath)
	if err != nil {
		return result // File doesn't exist or can't be read
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse "export NAME='value'" format
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimPrefix(line, "export ")
			if idx := strings.Index(line, "="); idx > 0 {
				name := line[:idx]
				value := line[idx+1:]

				// Remove only the outermost quotes (not all leading/trailing quote chars)
				// strings.Trim would remove ALL matching chars, corrupting escaped values
				if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
					value = value[1 : len(value)-1]
				}

				// Handle escaped single quotes: 'val'"'"'ue' -> val'ue
				// The pattern '\"'\"' is shell escaping for a literal single quote
				value = strings.ReplaceAll(value, "'\"'\"'", "'")

				result[name] = value
			}
		}
	}

	return result
}

// HasTmux checks if tmux is available in the container
func HasTmux(projectPath string) bool {
	_, err := execInContainer(projectPath, "which", "tmux")
	return err == nil
}

// KillTmuxSession kills a tmux session in the container
func KillTmuxSession(projectPath, sessionName string) error {
	return execInContainerWithStderr(projectPath, "failed to kill tmux session",
		"tmux", "kill-session", "-t", sessionName)
}
