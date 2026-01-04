package devcontainer

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/christophergyman/claude-quick/internal/constants"
)

// devcontainerFoundFunc is a callback invoked when a devcontainer.json is found
// configPath is the full path to devcontainer.json
// projectPath is the root directory of the project
type devcontainerFoundFunc func(configPath, projectPath string)

// walkDevcontainerDirs walks through search paths looking for devcontainer.json files
// and invokes the callback for each one found
func walkDevcontainerDirs(searchPaths []string, maxDepth int, excludedDirs []string, onFound devcontainerFoundFunc) {
	// Build exclusion set for O(1) lookup
	excludeSet := make(map[string]bool, len(excludedDirs))
	for _, dir := range excludedDirs {
		excludeSet[dir] = true
	}

	for _, searchPath := range searchPaths {
		filepath.WalkDir(searchPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // Skip directories we can't read
			}

			// Skip hidden directories (except .devcontainer)
			if d.IsDir() && strings.HasPrefix(d.Name(), ".") && d.Name() != constants.DevcontainerDir {
				return fs.SkipDir
			}

			// Skip excluded directories
			if d.IsDir() && excludeSet[d.Name()] {
				return fs.SkipDir
			}

			// Check depth
			relPath, _ := filepath.Rel(searchPath, path)
			depth := strings.Count(relPath, string(os.PathSeparator))
			if depth > maxDepth {
				return fs.SkipDir
			}

			// Look for devcontainer.json
			if d.Name() == constants.DevcontainerConfigFile {
				dir := filepath.Dir(path)

				// Only accept .devcontainer/devcontainer.json pattern
				if filepath.Base(dir) == constants.DevcontainerDir {
					projectPath := filepath.Dir(dir)
					onFound(path, projectPath)
				}
			}

			return nil
		})
	}
}

// DiscoverInstances finds all devcontainer instances in the given search paths
// For each project with a devcontainer.json, it finds all git worktrees
// and adds each worktree as a separate instance
func DiscoverInstances(searchPaths []string, maxDepth int, excludedDirs []string) []ContainerInstance {
	var instances []ContainerInstance
	seenProjects := make(map[string]bool)  // Track main repos we've processed
	seenWorktrees := make(map[string]bool) // Track worktree paths to deduplicate

	walkDevcontainerDirs(searchPaths, maxDepth, excludedDirs, func(configPath, projectPath string) {
		// Check if this is a git repo/worktree
		wtInfo := IsGitWorktree(projectPath)
		if wtInfo == nil {
			// Not a git repo - just add as a single instance without worktree info
			if !seenWorktrees[projectPath] {
				seenWorktrees[projectPath] = true
				instances = append(instances, ContainerInstance{
					Project: Project{
						Name: filepath.Base(projectPath),
						Path: projectPath,
					},
					ConfigPath: configPath,
					Worktree:   nil,
				})
			}
			return
		}

		// Get the main repo path
		mainRepo := wtInfo.MainRepo
		if seenProjects[mainRepo] {
			return // Already processed this project and its worktrees
		}
		seenProjects[mainRepo] = true

		// Find devcontainer.json in main repo (for worktrees to share)
		mainConfigPath := filepath.Join(mainRepo, ".devcontainer", "devcontainer.json")
		if _, err := os.Stat(mainConfigPath); err != nil {
			// Fall back to discovered config path
			mainConfigPath = configPath
		}

		// List all worktrees for this repository
		worktrees, err := ListWorktrees(mainRepo)
		if err != nil {
			// If we can't list worktrees, just add the discovered path
			if !seenWorktrees[projectPath] {
				seenWorktrees[projectPath] = true
				instances = append(instances, ContainerInstance{
					Project: Project{
						Name: filepath.Base(projectPath),
						Path: projectPath,
					},
					ConfigPath: mainConfigPath,
					Worktree:   wtInfo,
				})
			}
			return
		}

		// Add each worktree as a separate instance
		for _, wt := range worktrees {
			if seenWorktrees[wt.Path] {
				continue
			}
			seenWorktrees[wt.Path] = true

			// Copy worktree info
			wtCopy := wt
			instances = append(instances, ContainerInstance{
				Project: Project{
					Name: filepath.Base(mainRepo), // Use main repo name for all
					Path: wt.Path,
				},
				ConfigPath: mainConfigPath,
				Worktree:   &wtCopy,
			})
		}
	})

	return instances
}
