package devcontainer

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// Project represents a devcontainer project
type Project struct {
	Name string // Project directory name
	Path string // Full path to the project (workspace folder)
}

// Discover finds all devcontainer projects in the given search paths
func Discover(searchPaths []string, maxDepth int) []Project {
	var projects []Project
	seen := make(map[string]bool)

	for _, searchPath := range searchPaths {
		filepath.WalkDir(searchPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // Skip directories we can't read
			}

			// Skip hidden directories (except .devcontainer)
			if d.IsDir() && strings.HasPrefix(d.Name(), ".") && d.Name() != ".devcontainer" {
				return fs.SkipDir
			}

			// Skip node_modules and other common large directories
			if d.IsDir() {
				switch d.Name() {
				case "node_modules", "vendor", ".git", "__pycache__", "venv", ".venv":
					return fs.SkipDir
				}
			}

			// Check depth
			relPath, _ := filepath.Rel(searchPath, path)
			depth := strings.Count(relPath, string(os.PathSeparator))
			if depth > maxDepth {
				return fs.SkipDir
			}

			// Look for devcontainer.json
			if d.Name() == "devcontainer.json" {
				var projectPath string

				// Determine the project root
				dir := filepath.Dir(path)
				if filepath.Base(dir) == ".devcontainer" {
					// .devcontainer/devcontainer.json
					projectPath = filepath.Dir(dir)
				} else {
					// devcontainer.json at project root
					projectPath = dir
				}

				if !seen[projectPath] {
					seen[projectPath] = true
					projects = append(projects, Project{
						Name: filepath.Base(projectPath),
						Path: projectPath,
					})
				}
			}

			return nil
		})
	}

	return projects
}

// CheckCLI verifies the devcontainer CLI is installed
func CheckCLI() error {
	_, err := exec.LookPath("devcontainer")
	if err != nil {
		return fmt.Errorf("devcontainer CLI not found. Install with: npm install -g @devcontainers/cli")
	}
	return nil
}

// Up starts the devcontainer for a project
// Returns error if it fails
func Up(projectPath string) error {
	cmd := exec.Command("devcontainer", "up", "--workspace-folder", projectPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start container: %s", stderr.String())
	}
	return nil
}

// Stop stops the devcontainer by finding and stopping its Docker container
func Stop(projectPath string) error {
	// Find container by devcontainer label
	findCmd := exec.Command("docker", "ps", "-q",
		"--filter", fmt.Sprintf("label=devcontainer.local_folder=%s", projectPath))
	output, err := findCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to find container: %v", err)
	}
	containerID := strings.TrimSpace(string(output))
	if containerID == "" {
		return fmt.Errorf("no running container found for project")
	}
	// Stop the container
	stopCmd := exec.Command("docker", "stop", containerID)
	var stderr bytes.Buffer
	stopCmd.Stderr = &stderr
	if err := stopCmd.Run(); err != nil {
		return fmt.Errorf("failed to stop container: %s", stderr.String())
	}
	return nil
}

// Restart restarts the devcontainer
func Restart(projectPath string) error {
	findCmd := exec.Command("docker", "ps", "-a", "-q",
		"--filter", fmt.Sprintf("label=devcontainer.local_folder=%s", projectPath))
	output, err := findCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to find container: %v", err)
	}
	containerID := strings.TrimSpace(string(output))
	if containerID == "" {
		return Up(projectPath) // No container, just start
	}
	restartCmd := exec.Command("docker", "restart", containerID)
	var stderr bytes.Buffer
	restartCmd.Stderr = &stderr
	if err := restartCmd.Run(); err != nil {
		return fmt.Errorf("failed to restart container: %s", stderr.String())
	}
	return nil
}

// ExecInteractive executes a command inside the devcontainer interactively
// This replaces the current process with the devcontainer exec
func ExecInteractive(projectPath string, args []string) error {
	devcontainerPath, err := exec.LookPath("devcontainer")
	if err != nil {
		return err
	}

	cmdArgs := []string{"devcontainer", "exec", "--workspace-folder", projectPath}
	cmdArgs = append(cmdArgs, args...)

	// Replace current process with devcontainer exec
	return syscall.Exec(devcontainerPath, cmdArgs, os.Environ())
}

// ListTmuxSessions lists tmux sessions inside the container
func ListTmuxSessions(projectPath string) ([]string, error) {
	cmd := exec.Command("devcontainer", "exec", "--workspace-folder", projectPath,
		"tmux", "list-sessions", "-F", "#{session_name}:#{session_attached}")

	output, err := cmd.Output()
	if err != nil {
		// No sessions is not an error
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return nil, nil
			}
		}
		return nil, err
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
	cmd := exec.Command("devcontainer", "exec", "--workspace-folder", projectPath,
		"tmux", "new-session", "-d", "-s", sessionName)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create tmux session: %s", stderr.String())
	}
	return nil
}

// HasTmux checks if tmux is available in the container
func HasTmux(projectPath string) bool {
	cmd := exec.Command("devcontainer", "exec", "--workspace-folder", projectPath,
		"which", "tmux")
	return cmd.Run() == nil
}
