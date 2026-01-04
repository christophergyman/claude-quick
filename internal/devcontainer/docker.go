package devcontainer

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
)

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
	args := []string{"up", "--workspace-folder", projectPath}

	// For worktrees, mount the main repo's .git directory at the expected host path
	// This allows git to find the gitdir referenced in the worktree's .git file
	wtInfo := IsGitWorktree(projectPath)
	if wtInfo != nil && !wtInfo.IsMain {
		mainGitDir := filepath.Join(wtInfo.MainRepo, ".git")
		args = append(args, "--mount",
			fmt.Sprintf("type=bind,source=%s,target=%s", mainGitDir, mainGitDir))
	}

	cmd := exec.Command("devcontainer", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start container: %s", stderr.String())
	}
	return nil
}

// findContainerByPath finds a Docker container by its devcontainer.local_folder label
// If runningOnly is true, only searches running containers
// If runningOnly is false, searches all containers (including stopped)
func findContainerByPath(projectPath string, runningOnly bool) (string, error) {
	args := []string{"ps", "-q", "--filter", fmt.Sprintf("label=devcontainer.local_folder=%s", projectPath)}
	if !runningOnly {
		// Insert "-a" after "ps" to include stopped containers
		args = []string{"ps", "-a", "-q", "--filter", fmt.Sprintf("label=devcontainer.local_folder=%s", projectPath)}
	}
	cmd := exec.Command("docker", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to find container: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// Stop stops the devcontainer by finding and stopping its Docker container
func Stop(projectPath string) error {
	containerID, err := findContainerByPath(projectPath, true)
	if err != nil {
		return err
	}
	if containerID == "" {
		return fmt.Errorf("no running container found for project")
	}
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
	containerID, err := findContainerByPath(projectPath, false)
	if err != nil {
		return err
	}
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

// GetContainerStatus checks if a container is running for the given project path
func GetContainerStatus(projectPath string) (ContainerStatus, string) {
	// Check running containers first
	containerID, err := findContainerByPath(projectPath, true)
	if err != nil {
		return StatusUnknown, ""
	}
	if containerID != "" {
		return StatusRunning, containerID
	}

	// Check stopped containers
	cmd := exec.Command("docker", "ps", "-a", "-q",
		"--filter", fmt.Sprintf("label=devcontainer.local_folder=%s", projectPath),
		"--filter", "status=exited")
	output, err := cmd.Output()
	if err != nil {
		return StatusUnknown, ""
	}

	containerID = strings.TrimSpace(string(output))
	if containerID != "" {
		return StatusStopped, containerID
	}

	return StatusUnknown, ""
}

// GetAllInstancesStatus returns all instances with their current Docker status
func GetAllInstancesStatus(instances []ContainerInstance) []ContainerInstanceWithStatus {
	result := make([]ContainerInstanceWithStatus, len(instances))
	var wg sync.WaitGroup

	for i, inst := range instances {
		wg.Add(1)
		go func(idx int, instance ContainerInstance) {
			defer wg.Done()

			// Use path-based status check since each worktree has a unique path
			status, containerID := GetContainerStatus(instance.Path)
			sessionCount := 0

			// Only count sessions if container is running
			if status == StatusRunning {
				sessions, err := ListTmuxSessions(instance.Path)
				if err == nil {
					sessionCount = len(sessions)
				}
			}

			result[idx] = ContainerInstanceWithStatus{
				ContainerInstance: instance,
				Status:            status,
				ContainerID:       containerID,
				SessionCount:      sessionCount,
			}
		}(i, inst)
	}

	wg.Wait()
	return result
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

// execInContainer runs a command inside the devcontainer and returns its output
func execInContainer(projectPath string, args ...string) ([]byte, error) {
	cmdArgs := append([]string{"exec", "--workspace-folder", projectPath}, args...)
	cmd := exec.Command("devcontainer", cmdArgs...)
	return cmd.Output()
}

// execInContainerWithStderr runs a command inside the devcontainer and captures stderr for errors
func execInContainerWithStderr(projectPath string, errPrefix string, args ...string) error {
	cmdArgs := append([]string{"exec", "--workspace-folder", projectPath}, args...)
	cmd := exec.Command("devcontainer", cmdArgs...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", errPrefix, stderr.String())
	}
	return nil
}
