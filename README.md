# Claude Quick

TUI for orchestrating multiple Claude Code agents in devcontainers.

![Claude Quick Demo](claude-quick-nozoom.gif)

## Features

- **Unified Dashboard** - Discover and manage all your devcontainers from one place
- **Git Worktree Isolation** - Work on multiple branches in separate containers simultaneously
- **Credential Injection** - Securely pass API keys and tokens into containers
- **Interactive Wizard** - Guided setup on first run, no manual config required

## Prerequisites

- Go 1.25+
- Docker
- [devcontainer CLI](https://github.com/devcontainers/cli) (`npm install -g @devcontainers/cli`)
- tmux (inside your devcontainers)

## Quickstart

```bash
# Install
go install github.com/christophergyman/claude-quick@latest

# Run
claude-quick
```

The setup wizard launches automatically on first run. Follow the prompts to configure your search paths, credentials, and settings.

> **Tip:** Press `w` anytime to re-open the wizard and modify your configuration.

## Configuration

Configuration is stored in `claude-quick.yaml` next to the executable. The wizard handles this for you, but you can also edit it manually.

See [`claude-quick.yaml.example`](claude-quick.yaml.example) for all available options.

<details>
<summary><strong>Keybindings</strong></summary>

| Key | Action |
|-----|--------|
| `j`/`k` or `↑`/`↓` | Navigate |
| `Enter` | Select / Connect |
| `x` | Stop container or session |
| `r` | Restart |
| `R` | Refresh status |
| `w` | Open setup wizard |
| `n` | New worktree |
| `d` | Delete worktree |
| `?` | Show config |
| `q` / `Esc` | Back / Quit |

</details>

<details>
<summary><strong>Authentication</strong></summary>

Pipe credentials (API keys, tokens) into your devcontainers automatically.

Add credentials via the wizard, or manually in your config:

```yaml
auth:
  credentials:
    # Read from a file
    - name: ANTHROPIC_API_KEY
      source: file
      value: ~/.claude/.credentials

    # Read from environment variable
    - name: GITHUB_TOKEN
      source: env
      value: GITHUB_TOKEN

    # Run a command (e.g., password manager)
    - name: OPENAI_API_KEY
      source: command
      value: "op read op://Private/OpenAI/credential"
```

**Source types:**
| Type | Description | Value |
|------|-------------|-------|
| `file` | Read from a file | Path (supports `~`) |
| `env` | Read from host env var | Variable name |
| `command` | Run a command | Shell command |

Credentials are injected as environment variables in your tmux session and cleaned up when the container stops.

</details>

<details>
<summary><strong>Git Worktrees</strong></summary>

Each git worktree is treated as a separate devcontainer instance:

- **Create**: Press `n` on any git repository
- **Delete**: Press `d` to remove a worktree (stops container first)
- **View**: Worktrees appear as `project [branch-name]` in the dashboard

Constraints:
- Can only create worktrees on git repositories
- Cannot delete the main worktree

</details>

## Project Structure

```
claude-quick/
├── main.go              # Entry point
├── internal/
│   ├── config/          # YAML config loading
│   ├── auth/            # Credential management
│   ├── devcontainer/    # Container and git operations
│   └── tui/             # Terminal interface (Bubble Tea)
```

## Contributing

PRs welcome! Check out the [open issues](https://github.com/christophergyman/claude-quick/issues).

## License

MIT
