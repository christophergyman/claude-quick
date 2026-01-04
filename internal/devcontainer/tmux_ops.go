package devcontainer

import (
	"fmt"
	"os/exec"
	"strings"
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
	return execInContainerWithStderr(projectPath, "failed to create tmux session",
		"tmux", "new-session", "-d", "-s", sessionName)
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
