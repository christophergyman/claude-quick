package devcontainer

import (
	"path/filepath"
	"testing"
)

func TestNewMainWorktreeInfo(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		branch         string
		expectPath     string
		expectBranch   string
		expectMainRepo string
		expectGitDir   string
		expectIsMain   bool
	}{
		{
			name:           "basic main worktree",
			path:           "/home/user/projects/myapp",
			branch:         "main",
			expectPath:     "/home/user/projects/myapp",
			expectBranch:   "main",
			expectMainRepo: "/home/user/projects/myapp",
			expectGitDir:   filepath.Join("/home/user/projects/myapp", ".git"),
			expectIsMain:   true,
		},
		{
			name:           "master branch",
			path:           "/workspace/project",
			branch:         "master",
			expectPath:     "/workspace/project",
			expectBranch:   "master",
			expectMainRepo: "/workspace/project",
			expectGitDir:   filepath.Join("/workspace/project", ".git"),
			expectIsMain:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := newMainWorktreeInfo(tt.path, tt.branch)

			if result.Path != tt.expectPath {
				t.Errorf("Path = %q, want %q", result.Path, tt.expectPath)
			}
			if result.Branch != tt.expectBranch {
				t.Errorf("Branch = %q, want %q", result.Branch, tt.expectBranch)
			}
			if result.MainRepo != tt.expectMainRepo {
				t.Errorf("MainRepo = %q, want %q", result.MainRepo, tt.expectMainRepo)
			}
			if result.GitDir != tt.expectGitDir {
				t.Errorf("GitDir = %q, want %q", result.GitDir, tt.expectGitDir)
			}
			if result.IsMain != tt.expectIsMain {
				t.Errorf("IsMain = %v, want %v", result.IsMain, tt.expectIsMain)
			}
		})
	}
}

func TestNewBranchWorktreeInfo(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		branch         string
		mainRepo       string
		expectPath     string
		expectBranch   string
		expectMainRepo string
		expectGitDir   string
		expectIsMain   bool
	}{
		{
			name:           "feature branch worktree",
			path:           "/home/user/projects/myapp-feature",
			branch:         "feature/auth",
			mainRepo:       "/home/user/projects/myapp",
			expectPath:     "/home/user/projects/myapp-feature",
			expectBranch:   "feature/auth",
			expectMainRepo: "/home/user/projects/myapp",
			expectGitDir:   filepath.Join("/home/user/projects/myapp", ".git", "worktrees", "myapp-feature"),
			expectIsMain:   false,
		},
		{
			name:           "bugfix worktree",
			path:           "/workspace/project-fix-123",
			branch:         "bugfix/issue-123",
			mainRepo:       "/workspace/project",
			expectPath:     "/workspace/project-fix-123",
			expectBranch:   "bugfix/issue-123",
			expectMainRepo: "/workspace/project",
			expectGitDir:   filepath.Join("/workspace/project", ".git", "worktrees", "project-fix-123"),
			expectIsMain:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := newBranchWorktreeInfo(tt.path, tt.branch, tt.mainRepo)

			if result.Path != tt.expectPath {
				t.Errorf("Path = %q, want %q", result.Path, tt.expectPath)
			}
			if result.Branch != tt.expectBranch {
				t.Errorf("Branch = %q, want %q", result.Branch, tt.expectBranch)
			}
			if result.MainRepo != tt.expectMainRepo {
				t.Errorf("MainRepo = %q, want %q", result.MainRepo, tt.expectMainRepo)
			}
			if result.GitDir != tt.expectGitDir {
				t.Errorf("GitDir = %q, want %q", result.GitDir, tt.expectGitDir)
			}
			if result.IsMain != tt.expectIsMain {
				t.Errorf("IsMain = %v, want %v", result.IsMain, tt.expectIsMain)
			}
		})
	}
}

func TestContainerInstance_DisplayName(t *testing.T) {
	tests := []struct {
		name     string
		instance ContainerInstance
		expected string
	}{
		{
			name: "without worktree",
			instance: ContainerInstance{
				Project: Project{Name: "myapp", Path: "/path/to/myapp"},
			},
			expected: "myapp",
		},
		{
			name: "with main worktree",
			instance: ContainerInstance{
				Project: Project{Name: "myapp", Path: "/path/to/myapp"},
				Worktree: &WorktreeInfo{
					Path:   "/path/to/myapp",
					Branch: "main",
					IsMain: true,
				},
			},
			expected: "myapp",
		},
		{
			name: "with feature branch worktree",
			instance: ContainerInstance{
				Project: Project{Name: "myapp", Path: "/path/to/myapp-feature"},
				Worktree: &WorktreeInfo{
					Path:   "/path/to/myapp-feature",
					Branch: "feature/auth",
					IsMain: false,
				},
			},
			expected: "myapp [feature/auth]",
		},
		{
			name: "with bugfix branch worktree",
			instance: ContainerInstance{
				Project: Project{Name: "project", Path: "/workspace/project-fix"},
				Worktree: &WorktreeInfo{
					Path:   "/workspace/project-fix",
					Branch: "bugfix/issue-42",
					IsMain: false,
				},
			},
			expected: "project [bugfix/issue-42]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.instance.DisplayName()
			if result != tt.expected {
				t.Errorf("DisplayName() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestContainerStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		status   ContainerStatus
		expected string
	}{
		{"running", StatusRunning, "running"},
		{"stopped", StatusStopped, "stopped"},
		{"unknown", StatusUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("ContainerStatus = %q, want %q", tt.status, tt.expected)
			}
		})
	}
}

func TestProject(t *testing.T) {
	p := Project{
		Name: "test-project",
		Path: "/home/user/projects/test-project",
	}

	if p.Name != "test-project" {
		t.Errorf("Project.Name = %q, want %q", p.Name, "test-project")
	}
	if p.Path != "/home/user/projects/test-project" {
		t.Errorf("Project.Path = %q, want %q", p.Path, "/home/user/projects/test-project")
	}
}

func TestContainerInstanceWithStatus(t *testing.T) {
	instance := ContainerInstanceWithStatus{
		ContainerInstance: ContainerInstance{
			Project:    Project{Name: "myapp", Path: "/path/to/myapp"},
			ConfigPath: "/path/to/myapp/.devcontainer/devcontainer.json",
		},
		Status:       StatusRunning,
		ContainerID:  "abc123def456",
		SessionCount: 3,
	}

	if instance.Status != StatusRunning {
		t.Errorf("Status = %q, want %q", instance.Status, StatusRunning)
	}
	if instance.ContainerID != "abc123def456" {
		t.Errorf("ContainerID = %q, want %q", instance.ContainerID, "abc123def456")
	}
	if instance.SessionCount != 3 {
		t.Errorf("SessionCount = %d, want %d", instance.SessionCount, 3)
	}
	// Verify embedded ContainerInstance fields
	if instance.Name != "myapp" {
		t.Errorf("Name = %q, want %q", instance.Name, "myapp")
	}
}
