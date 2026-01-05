package devcontainer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWalkDevcontainerDirs_FindsDevcontainer(t *testing.T) {
	// Create a temp directory structure with a devcontainer
	tmpDir, err := os.MkdirTemp("", "test-discovery-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create project with devcontainer
	projectDir := filepath.Join(tmpDir, "myproject")
	devcontainerDir := filepath.Join(projectDir, ".devcontainer")
	if err := os.MkdirAll(devcontainerDir, 0755); err != nil {
		t.Fatalf("failed to create devcontainer dir: %v", err)
	}

	// Create devcontainer.json
	devcontainerJson := filepath.Join(devcontainerDir, "devcontainer.json")
	if err := os.WriteFile(devcontainerJson, []byte(`{"name": "test"}`), 0644); err != nil {
		t.Fatalf("failed to create devcontainer.json: %v", err)
	}

	// Test discovery
	var found []string
	walkDevcontainerDirs(
		[]string{tmpDir},
		3,
		[]string{},
		func(configPath, projectPath string) {
			found = append(found, projectPath)
		},
	)

	if len(found) != 1 {
		t.Errorf("expected 1 project, found %d", len(found))
	}
	if len(found) > 0 && found[0] != projectDir {
		t.Errorf("found project = %q, want %q", found[0], projectDir)
	}
}

func TestWalkDevcontainerDirs_SkipsExcludedDirs(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-discovery-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create project in node_modules (should be skipped)
	excludedProject := filepath.Join(tmpDir, "node_modules", "somepackage")
	devcontainerDir := filepath.Join(excludedProject, ".devcontainer")
	if err := os.MkdirAll(devcontainerDir, 0755); err != nil {
		t.Fatalf("failed to create devcontainer dir: %v", err)
	}
	devcontainerJson := filepath.Join(devcontainerDir, "devcontainer.json")
	if err := os.WriteFile(devcontainerJson, []byte(`{"name": "test"}`), 0644); err != nil {
		t.Fatalf("failed to create devcontainer.json: %v", err)
	}

	// Create normal project
	normalProject := filepath.Join(tmpDir, "normal")
	normalDevcontainer := filepath.Join(normalProject, ".devcontainer")
	if err := os.MkdirAll(normalDevcontainer, 0755); err != nil {
		t.Fatalf("failed to create devcontainer dir: %v", err)
	}
	normalJson := filepath.Join(normalDevcontainer, "devcontainer.json")
	if err := os.WriteFile(normalJson, []byte(`{"name": "normal"}`), 0644); err != nil {
		t.Fatalf("failed to create devcontainer.json: %v", err)
	}

	var found []string
	walkDevcontainerDirs(
		[]string{tmpDir},
		3,
		[]string{"node_modules"},
		func(configPath, projectPath string) {
			found = append(found, projectPath)
		},
	)

	// Should only find the normal project, not the one in node_modules
	if len(found) != 1 {
		t.Errorf("expected 1 project (excluded node_modules), found %d", len(found))
	}
	if len(found) > 0 && found[0] != normalProject {
		t.Errorf("found project = %q, want %q", found[0], normalProject)
	}
}

func TestWalkDevcontainerDirs_RespectsMaxDepth(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-discovery-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create shallow project (depth 1 from tmpDir)
	// Path: tmpDir/shallow/.devcontainer/devcontainer.json
	// Depth to devcontainer.json = 2 (shallow/.devcontainer/devcontainer.json has 2 separators)
	shallowProject := filepath.Join(tmpDir, "shallow")
	shallowDevcontainer := filepath.Join(shallowProject, ".devcontainer")
	if err := os.MkdirAll(shallowDevcontainer, 0755); err != nil {
		t.Fatalf("failed to create devcontainer dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(shallowDevcontainer, "devcontainer.json"), []byte(`{}`), 0644); err != nil {
		t.Fatalf("failed to create devcontainer.json: %v", err)
	}

	// Create deep project (depth 5 from tmpDir)
	// Path: tmpDir/level1/level2/deep/.devcontainer/devcontainer.json
	// Depth = 5 (level1/level2/deep/.devcontainer/devcontainer.json has 4 separators)
	deepProject := filepath.Join(tmpDir, "level1", "level2", "deep")
	deepDevcontainer := filepath.Join(deepProject, ".devcontainer")
	if err := os.MkdirAll(deepDevcontainer, 0755); err != nil {
		t.Fatalf("failed to create devcontainer dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(deepDevcontainer, "devcontainer.json"), []byte(`{}`), 0644); err != nil {
		t.Fatalf("failed to create devcontainer.json: %v", err)
	}

	// Test with maxDepth=2 (should find shallow but not deep)
	// shallow/.devcontainer/devcontainer.json has depth 2, so maxDepth=2 allows it
	// level1/level2/deep/.devcontainer/devcontainer.json has depth 4, maxDepth=2 blocks at level1/level2
	var found []string
	walkDevcontainerDirs(
		[]string{tmpDir},
		2,
		[]string{},
		func(configPath, projectPath string) {
			found = append(found, filepath.Base(projectPath))
		},
	)

	foundShallow := false
	foundDeep := false
	for _, name := range found {
		if name == "shallow" {
			foundShallow = true
		}
		if name == "deep" {
			foundDeep = true
		}
	}

	if !foundShallow {
		t.Error("expected to find shallow project at maxDepth=2")
	}
	if foundDeep {
		t.Error("should not find deep project at maxDepth=2")
	}
}

func TestWalkDevcontainerDirs_SkipsHiddenDirs(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-discovery-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create project in hidden directory (should be skipped)
	hiddenProject := filepath.Join(tmpDir, ".hidden", "project")
	hiddenDevcontainer := filepath.Join(hiddenProject, ".devcontainer")
	if err := os.MkdirAll(hiddenDevcontainer, 0755); err != nil {
		t.Fatalf("failed to create devcontainer dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(hiddenDevcontainer, "devcontainer.json"), []byte(`{}`), 0644); err != nil {
		t.Fatalf("failed to create devcontainer.json: %v", err)
	}

	// Create normal project
	normalProject := filepath.Join(tmpDir, "visible")
	normalDevcontainer := filepath.Join(normalProject, ".devcontainer")
	if err := os.MkdirAll(normalDevcontainer, 0755); err != nil {
		t.Fatalf("failed to create devcontainer dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(normalDevcontainer, "devcontainer.json"), []byte(`{}`), 0644); err != nil {
		t.Fatalf("failed to create devcontainer.json: %v", err)
	}

	var found []string
	walkDevcontainerDirs(
		[]string{tmpDir},
		3,
		[]string{},
		func(configPath, projectPath string) {
			found = append(found, filepath.Base(projectPath))
		},
	)

	// Should only find visible project, not the one in .hidden
	if len(found) != 1 {
		t.Errorf("expected 1 project, found %d: %v", len(found), found)
	}
	if len(found) > 0 && found[0] != "visible" {
		t.Errorf("found project = %q, want 'visible'", found[0])
	}
}

func TestWalkDevcontainerDirs_MultipleSearchPaths(t *testing.T) {
	// Create two separate temp directories
	tmpDir1, err := os.MkdirTemp("", "test-discovery1-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir1)

	tmpDir2, err := os.MkdirTemp("", "test-discovery2-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir2)

	// Create project in first directory
	project1 := filepath.Join(tmpDir1, "project1")
	if err := os.MkdirAll(filepath.Join(project1, ".devcontainer"), 0755); err != nil {
		t.Fatalf("failed to create devcontainer dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(project1, ".devcontainer", "devcontainer.json"), []byte(`{}`), 0644); err != nil {
		t.Fatalf("failed to create devcontainer.json: %v", err)
	}

	// Create project in second directory
	project2 := filepath.Join(tmpDir2, "project2")
	if err := os.MkdirAll(filepath.Join(project2, ".devcontainer"), 0755); err != nil {
		t.Fatalf("failed to create devcontainer dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(project2, ".devcontainer", "devcontainer.json"), []byte(`{}`), 0644); err != nil {
		t.Fatalf("failed to create devcontainer.json: %v", err)
	}

	var found []string
	walkDevcontainerDirs(
		[]string{tmpDir1, tmpDir2},
		3,
		[]string{},
		func(configPath, projectPath string) {
			found = append(found, filepath.Base(projectPath))
		},
	)

	// Should find both projects
	if len(found) != 2 {
		t.Errorf("expected 2 projects, found %d: %v", len(found), found)
	}

	foundProject1 := false
	foundProject2 := false
	for _, name := range found {
		if name == "project1" {
			foundProject1 = true
		}
		if name == "project2" {
			foundProject2 = true
		}
	}

	if !foundProject1 {
		t.Error("expected to find project1")
	}
	if !foundProject2 {
		t.Error("expected to find project2")
	}
}

func TestWalkDevcontainerDirs_EmptySearchPaths(t *testing.T) {
	var found []string
	walkDevcontainerDirs(
		[]string{},
		3,
		[]string{},
		func(configPath, projectPath string) {
			found = append(found, projectPath)
		},
	)

	if len(found) != 0 {
		t.Errorf("expected 0 projects with empty search paths, found %d", len(found))
	}
}

func TestWalkDevcontainerDirs_NonexistentSearchPath(t *testing.T) {
	var found []string
	walkDevcontainerDirs(
		[]string{"/nonexistent/path/that/does/not/exist"},
		3,
		[]string{},
		func(configPath, projectPath string) {
			found = append(found, projectPath)
		},
	)

	// Should handle gracefully without error
	if len(found) != 0 {
		t.Errorf("expected 0 projects for nonexistent path, found %d", len(found))
	}
}

func TestDiscoverInstances_NonGitProject(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-discover-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create project without git (no .git directory)
	project := filepath.Join(tmpDir, "non-git-project")
	devcontainer := filepath.Join(project, ".devcontainer")
	if err := os.MkdirAll(devcontainer, 0755); err != nil {
		t.Fatalf("failed to create devcontainer dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(devcontainer, "devcontainer.json"), []byte(`{}`), 0644); err != nil {
		t.Fatalf("failed to create devcontainer.json: %v", err)
	}

	instances := DiscoverInstances([]string{tmpDir}, 3, []string{})

	if len(instances) != 1 {
		t.Errorf("expected 1 instance, got %d", len(instances))
	}
	if len(instances) > 0 {
		if instances[0].Name != "non-git-project" {
			t.Errorf("Name = %q, want 'non-git-project'", instances[0].Name)
		}
		if instances[0].Worktree != nil {
			t.Error("Worktree should be nil for non-git project")
		}
	}
}

func TestDiscoverInstances_GitProject(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-discover-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create project with git (.git directory)
	project := filepath.Join(tmpDir, "git-project")
	devcontainer := filepath.Join(project, ".devcontainer")
	gitDir := filepath.Join(project, ".git")
	if err := os.MkdirAll(devcontainer, 0755); err != nil {
		t.Fatalf("failed to create devcontainer dir: %v", err)
	}
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(devcontainer, "devcontainer.json"), []byte(`{}`), 0644); err != nil {
		t.Fatalf("failed to create devcontainer.json: %v", err)
	}

	instances := DiscoverInstances([]string{tmpDir}, 3, []string{})

	// Should find at least 1 instance (the main repo)
	if len(instances) == 0 {
		t.Error("expected at least 1 instance")
	}
	if len(instances) > 0 {
		// The instance should have worktree info for a git repo
		if instances[0].Worktree == nil {
			t.Log("Worktree info may be nil if git commands fail in test environment")
		} else if !instances[0].Worktree.IsMain {
			t.Error("Expected main worktree to have IsMain=true")
		}
	}
}

func TestDiscoverInstances_Deduplication(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-discover-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a single project
	project := filepath.Join(tmpDir, "project")
	devcontainer := filepath.Join(project, ".devcontainer")
	if err := os.MkdirAll(devcontainer, 0755); err != nil {
		t.Fatalf("failed to create devcontainer dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(devcontainer, "devcontainer.json"), []byte(`{}`), 0644); err != nil {
		t.Fatalf("failed to create devcontainer.json: %v", err)
	}

	// Search from multiple paths that would find the same project
	instances := DiscoverInstances([]string{tmpDir, tmpDir}, 3, []string{})

	// Should not have duplicates
	seen := make(map[string]bool)
	for _, inst := range instances {
		if seen[inst.Path] {
			t.Errorf("duplicate instance found: %s", inst.Path)
		}
		seen[inst.Path] = true
	}
}
