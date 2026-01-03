# Claude Quick

A terminal user interface for managing tmux sessions inside devcontainers.

Claude Quick discovers devcontainer projects on your system, spins up containers, and provides an intuitive interface for creating and attaching to tmux sessions within them.

## Features

- **Project Discovery** - Automatically finds devcontainer projects across configured search paths
- **Container Management** - Start, stop, and restart devcontainers using the official CLI
- **Tmux Integration** - List, create, stop, restart, and attach to tmux sessions inside containers
- **Keyboard Navigation** - Vim-style keybindings (j/k) and arrow key support
- **Configurable** - YAML configuration for search paths and scan depth

## Requirements

- Go 1.21+ (for building)
- Docker
- [devcontainer CLI](https://github.com/devcontainers/cli) - Install with:
  ```bash
  npm install -g @devcontainers/cli
  ```
- tmux installed inside your devcontainers

## Installation

### Build from Source

```bash
git clone https://github.com/christophergyman/claude-quick.git
cd claude-quick
go build -o claude-quick .
```

Optionally, add to your PATH by creating a symlink:

```bash
mkdir -p ~/.local/bin
ln -sf "$(pwd)/claude-quick" ~/.local/bin/claude-quick
```

### Go Install

```bash
go install github.com/christophergyman/claude-quick@latest
```

## Usage

Run the application:

```bash
claude-quick
```

### Workflow

1. Select a devcontainer project from the discovered list
2. Wait for the container to start
3. Choose an existing tmux session or create a new one
4. Claude Quick attaches you directly to the tmux session inside the container

### Keybindings

#### Container Selection

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `Enter` | Select and start container |
| `x` | Stop container |
| `r` | Restart container |
| `q` | Quit |
| `Ctrl+C` | Quit |

#### Tmux Session Selection

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `Enter` | Select and attach to session |
| `x` | Stop (kill) session |
| `r` | Restart session |
| `q` / `Esc` | Go back to container list |
| `Ctrl+C` | Quit |

#### New Session Input

| Key | Action |
|-----|--------|
| `Enter` | Create session with entered name |
| `Esc` | Cancel and go back |
| `Ctrl+C` | Quit |

#### Confirmation Dialogs

| Key | Action |
|-----|--------|
| `y` | Confirm action |
| `n` / `Esc` | Cancel |
| `Ctrl+C` | Quit |

## Configuration

Claude Quick looks for a configuration file at `~/.config/claude-quick/config.yaml`.

### Example Configuration

```yaml
# Directories to scan for devcontainer projects
search_paths:
  - ~/projects
  - ~/work
  - ~/Documents/github

# Maximum directory depth to search (default: 3)
max_depth: 4

# Directories to skip during scanning (defaults shown below)
excluded_dirs:
  - node_modules
  - vendor
  - .git
  - __pycache__
  - venv
  - .venv
  - dist
  - build
  - .cache

# Default name for new tmux sessions (default: "main")
default_session_name: dev

# Container startup timeout in seconds (default: 300)
container_timeout_seconds: 300
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `search_paths` | list | `[~]` | Directories to scan for devcontainer projects |
| `max_depth` | int | `3` | Maximum directory depth for scanning |
| `excluded_dirs` | list | See below | Directories to skip during scanning |
| `default_session_name` | string | `main` | Default name for new tmux sessions |
| `container_timeout_seconds` | int | `300` | Timeout for container startup (30-1800) |

#### Default Excluded Directories

By default, the following directories are skipped during scanning:
`node_modules`, `vendor`, `.git`, `__pycache__`, `venv`, `.venv`, `dist`, `build`, `.cache`

### Default Behavior

Without a configuration file, Claude Quick searches your home directory with a max depth of 3.

## Nerd Font Support for Claude CLI

If you're using [Claude Code](https://claude.com/claude-code) inside your devcontainers, you'll need to configure your devcontainer for proper Nerd Font rendering. Without this, Unicode glyphs (like the Claude logo) won't display correctly.

Add the following to your `devcontainer.json`:

```json
{
  "containerEnv": {
    "TERM": "xterm-256color",
    "LANG": "en_US.UTF-8"
  }
}
```

**Why is this needed?** The terminal is rendered by your host machine (which has Nerd Fonts installed), but applications inside the container need `LANG=en_US.UTF-8` to know they can output UTF-8/Unicode characters.

## How It Works

1. Scans configured paths for `devcontainer.json` files (skips hidden directories, node_modules, vendor, etc.)
2. Uses `devcontainer up` to ensure the container is running
3. Checks that tmux is available in the container
4. Queries tmux inside the container for existing sessions
5. On selection, executes `devcontainer exec` to attach to the tmux session

## License

MIT
