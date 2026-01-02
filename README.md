# QuickVibe

A terminal user interface for managing tmux sessions inside devcontainers.

QuickVibe discovers devcontainer projects on your system, spins up containers, and provides an intuitive interface for creating and attaching to tmux sessions within them.

## Features

- **Project Discovery** - Automatically finds devcontainer projects across configured search paths
- **Container Management** - Starts devcontainers using the official CLI
- **Tmux Integration** - List, create, and attach to tmux sessions inside containers
- **Keyboard Navigation** - Vim-style keybindings (j/k) and arrow key support
- **Configurable** - YAML configuration for search paths and scan depth

## Prerequisites

- Go 1.21+
- [devcontainer CLI](https://github.com/devcontainers/cli) - Install with:
  ```bash
  npm install -g @devcontainers/cli
  ```
- tmux installed inside your devcontainers

## Installation

```bash
go install github.com/chezu/quickvibe@latest
```

Or build from source:

```bash
git clone https://github.com/chezu/quickvibe.git
cd quickvibe
go build -o quickvibe .
```

## Usage

Run the application:

```bash
quickvibe
```

### Workflow

1. Select a devcontainer project from the discovered list
2. Wait for the container to start
3. Choose an existing tmux session or create a new one
4. QuickVibe attaches you directly to the tmux session inside the container

### Keybindings

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `Enter` | Select |
| `Esc` / `q` | Go back / Quit |
| `Ctrl+C` | Quit |

## Configuration

QuickVibe looks for a configuration file at `~/.config/quickvibe/config.yaml`.

### Example Configuration

```yaml
# Directories to scan for devcontainer projects
search_paths:
  - ~/projects
  - ~/work
  - ~/Documents/github

# Maximum directory depth to search (default: 3)
max_depth: 4
```

### Default Behavior

Without a configuration file, QuickVibe searches your home directory with a max depth of 3.

## How It Works

1. Scans configured paths for `devcontainer.json` files
2. Uses `devcontainer up` to ensure the container is running
3. Queries tmux inside the container for existing sessions
4. On selection, executes `devcontainer exec` to attach to the tmux session

## License

MIT
