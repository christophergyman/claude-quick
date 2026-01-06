// Package devcontainer handles container lifecycle and git operations.
//
// # Overview
//
// This package provides functionality for:
//   - Discovering devcontainer projects in search paths
//   - Managing container lifecycle (start, stop, restart)
//   - Git worktree operations
//   - tmux session management within containers
//
// # Key Types
//
//   - ContainerInstance: Represents a discovered devcontainer project
//   - WorktreeInfo: Git worktree metadata (branch, path, main repo status)
//
// # Key Files
//
//   - discovery.go: Recursive devcontainer.json scanner
//   - docker.go: Container lifecycle (up, stop, restart, status checks)
//   - git.go: Worktree detection, creation, deletion, branch validation
//   - tmux_ops.go: Session management, credential injection
//   - types.go: Type definitions
//
// # Container Identification
//
// Uses Docker label queries for reliability:
//
//	docker ps --filter label=devcontainer.local_folder=<path>
//
// # Git Worktree Integration
//
// Each worktree is treated as a separate devcontainer instance.
// Worktrees share the same devcontainer.json from the main repo.
// The container mounts the main repo's .git directory for git operations.
package devcontainer
