# Claude Quick

A terminal UI (TUI) application written in Go that manages devcontainers with git worktrees and tmux sessions. It provides a unified dashboard for discovering, starting, and connecting to containerized development environments.

## Architecture Overview

```
User → TUI (Bubble Tea) → devcontainer CLI → Docker → tmux sessions
                       ↘ git worktree operations
```

**Core Flow**: Select project → Start container → Pick/create tmux session → Attach with credentials injected

## Project Structure

```
claude-quick/
├── main.go                    # Entry point: loads config, validates CLI, launches TUI
├── internal/
│   ├── config/config.go       # YAML config loading (executable dir or ~/.config/claude-quick/)
│   ├── constants/constants.go # Default values, timeouts, display limits
│   ├── auth/                  # Credential management
│   │   ├── types.go           # Credential, SourceType (file/env/command)
│   │   ├── resolver.go        # Multi-source credential resolution
│   │   └── file.go            # .claude-quick-auth file I/O
│   ├── devcontainer/          # Container and git operations
│   │   ├── types.go           # Project, WorktreeInfo, ContainerInstance
│   │   ├── discovery.go       # Recursive devcontainer.json scanner
│   │   ├── docker.go          # Container lifecycle (up/stop/restart)
│   │   ├── git.go             # Worktree detection, creation, deletion
│   │   └── tmux_ops.go        # Session management, credential injection
│   ├── tmux/tmux.go           # Session parsing utilities
│   └── tui/                   # Terminal interface
│       ├── model.go           # Bubble Tea model definition
│       ├── state.go           # State machine (23 states)
│       ├── handlers.go        # Keyboard event handlers
│       ├── commands.go        # Async commands (discovery, container ops)
│       ├── messages.go        # Message types for async results
│       ├── container.go       # Dashboard rendering
│       ├── tmux.go            # Session selection rendering
│       └── styles.go          # Lipgloss styling (purple theme)
```

## Key Concepts

### State Machine Pattern

The TUI operates as a deterministic state machine. Key states:
- `StateDashboard` - Main container list
- `StateContainerStarting/Stopping` - Container operations in progress
- `StateTmuxSelect` - Session picker after container starts
- `StateNewWorktreeInput` - Creating git worktree

All transitions are explicit via message handling in `Update()`.

### Async Command Pattern

No blocking I/O in the UI. Operations return `tea.Cmd` that execute async:
```go
discoverInstances() → instancesDiscoveredMsg
startContainer()    → containerStartedMsg
loadTmuxSessions()  → tmuxSessionsLoadedMsg
```

### Git Worktree Integration

Each worktree is treated as a separate devcontainer instance:
- Discovery finds all worktrees via `git worktree list --porcelain`
- Worktrees share the same `devcontainer.json` from main repo
- Container mounts main repo's `.git` directory for git operations
- Branch name shown in dashboard: `project [branch-name]`

### Credential Injection

Three credential sources:
1. `file` - Read from a file (e.g., `~/.claude/.credentials`)
2. `env` - Read from host environment variable
3. `command` - Execute command (e.g., 1Password CLI)

Flow: Resolve credentials → Write to `.claude-quick-auth` → Inject into tmux session

**Credentials are available as environment variables in the tmux session**. They are injected in two ways:
1. Via `tmux new-session -e NAME=value` - available to the initial shell immediately
2. Via `tmux setenv` - propagates to any new windows/panes created later

If you need to access credentials (e.g., `GITHUB_TOKEN` for git push), they should be available as normal environment variables. If not, you can source the auth file:
```bash
source /workspaces/<project>/.claude-quick-auth
```

## Important Implementation Details

### Container Identification

Uses Docker label queries for reliability:
```bash
docker ps --filter label=devcontainer.local_folder=<path>
```

### Worktree Container Mounting

For worktrees, the container needs access to the main repo's `.git`:
```go
// Bind mount: host main .git → container worktree expects it at host path
```

### Credential File Security

- Written with `0600` permissions
- Shell-compatible format: `export VAR='value'`
- Handles quote escaping: `value's` → `value'"'"'s`

### Status Checks

Parallel goroutines query Docker status for all instances simultaneously.

## Configuration

Location (in priority order):
1. `claude-quick.yaml` next to executable (following symlinks)
2. `~/.config/claude-quick/config.yaml` (legacy, deprecated)

```yaml
search_paths:
  - ~/projects
max_depth: 3
excluded_dirs: [node_modules, vendor, .git]
default_session_name: main
container_timeout_seconds: 300
launch_command: "claude"  # Command to run when a new tmux session is created

auth:
  credentials:
    - name: ANTHROPIC_API_KEY
      source: file
      value: ~/.claude/.credentials
  projects:
    my-project:
      launch_command: "npm run dev"  # Per-project override
      credentials:
        - name: API_KEY
          source: env
          value: MY_API_KEY

github:
  default_state: open              # Filter for issue list: open, closed, or all
  branch_prefix: "issue-"          # Prefix for auto-generated branch names
  max_issues: 50                   # Maximum issues to fetch
  in_progress_label: "in-progress" # Label added when creating worktree from issue
  label_color: "fbca04"            # Hex color for auto-created label (yellow)
  label_description: "Issue is being actively worked on"  # Description for label
  auto_label_issues: true          # Enable/disable auto-labeling (default: true)
  create_label_if_missing: true    # Auto-create label if missing (default: true)
```

## Keybindings

| Key | Action |
|-----|--------|
| `j`/`k`, `↑`/`↓` | Navigate |
| `Enter` | Select/Start |
| `x` | Stop container/session |
| `r` | Restart |
| `R` | Refresh status |
| `n` | New worktree |
| `d` | Delete worktree |
| `?` | Show config |
| `Esc`/`q` | Back/Quit |

## Dependencies

**Required on host**:
- `git` - Worktree operations
- `docker` - Container runtime
- `devcontainer` CLI - Container lifecycle (`npm install -g @devcontainers/cli`)

**Required in containers**:
- `tmux` - Session management

**Go dependencies**:
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/bubbles` - UI components
- `github.com/charmbracelet/lipgloss` - Styling
- `gopkg.in/yaml.v3` - Config parsing

## Build

```bash
go build -o claude-quick .
# Or use the build script:
./build.sh  # Builds and symlinks to ~/.local/bin
```

## Testing Changes

1. Run `go build` to verify compilation
2. Test discovery: Launch with projects in search paths
3. Test container ops: Start/stop containers
4. Test worktrees: Create/delete worktrees
5. Test tmux: Create/attach/kill sessions

## Common Gotchas

**Adding new UI states** requires changes in 3 places:
1. `tui/state.go` - Define the state constant
2. `tui/model.go` Update() - Handle state transitions
3. `tui/model.go` View() - Render the state

**Path vs Name**: Always use `instance.Path` for Docker/devcontainer operations, never `instance.Name`. Name is just the directory name (same for all worktrees), Path is the unique workspace path.

**Credential escaping**: `auth/file.go` (write) and `devcontainer/tmux_ops.go` (read) must stay in sync. Single quotes use `'val'"'"'ue'` pattern.

**Path expansion**: The `util.ExpandPath()` function in `internal/util/path.go` handles `~` expansion. Used by both config and auth packages.

## State Machine Rules

- States defined via `iota` in `tui/state.go` - order is significant
- Transitions are generally one-way; going back requires explicit handlers
- Always set `m.state` BEFORE returning from Update()
- Text inputs (`textInput`, `worktreeInput`) only receive keystrokes when in their respective states (`StateNewSessionInput`, `StateNewWorktreeInput`)
- After async operations complete, the message handler MUST transition to a specific next state

## Cursor/Selection Invariants

- Use `TotalTmuxOptions()` helper when checking tmux list bounds (accounts for "+New Session" option)
- Use `IsNewSessionSelected()` instead of raw `cursor == len(sessions)` comparison
- Always bounds-check `m.cursor` before accessing `m.instancesStatus[m.cursor]`
- Cursor is reused across contexts (instances, sessions) - reset appropriately on state transitions

## Container Startup Order

Operations must happen in this sequence:
1. Resolve credentials via `auth.Resolve()`
2. Write `.claude-quick-auth` file (must exist before container exec)
3. Run `devcontainer up --workspace-folder <path>`
4. Check tmux exists in container (hard failure if missing)
5. Load tmux sessions

On container stop:
- Call `auth.CleanupCredentialFile()` to remove `.claude-quick-auth`

## Worktree Mount Binding

For git worktrees, the `.git` is a FILE containing a path like `gitdir: /main/repo/.git/worktrees/branch`.

When starting a worktree container, we bind-mount the main repo's `.git` directory:
```go
--mount type=bind,source=/main/repo/.git,target=/main/repo/.git
```

**Critical**: Source and target paths MUST be identical. The worktree's `.git` file contains an absolute host path that must remain valid inside the container.

## Goroutine Pattern

The only goroutine usage is in `docker.go` GetAllInstancesStatus():
```go
go func(idx int, instance ContainerInstance) {
    // ...
}(i, inst)  // Pass as parameters, don't capture loop variables
```

This pattern avoids the classic Go loop variable capture bug. Always pass loop variables as function parameters when spawning goroutines.

## Testing

Run tests with:
```bash
go test ./...
```

Test coverage exists for:
- `internal/auth` - Credential resolution, file operations, quote escaping
- `internal/config` - Configuration loading, validation, defaults
- `internal/constants` - Constant values
- `internal/devcontainer` - Discovery, git worktrees, depth limits
- `internal/tui` - Helpers, styles, rendering functions, model accessors
- `internal/util` - Path expansion

The build script (`./build.sh`) runs tests automatically before building.
