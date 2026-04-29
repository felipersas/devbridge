# NotifyBridge

Send notifications from macOS to Android via SSH/Tailscale. Built for Claude Code hooks, terminal workflows, and remote development.

## Features

- Push notifications from Mac to Android phone
- SSH/Tailscale transport (no cloud service needed)
- QR code pairing with bidirectional SSH key exchange
- Claude Code hook integration (auto-notify when Claude finishes)
- Per-project profiles (custom emoji, LED color, priority)
- `claude-remote` command on Android (SSH into Mac tmux+Claude session)
- Retry with configurable delay
- Interactive setup wizard

## Quick Start

```bash
# Build and install
make build
make install

# Pair with your Android phone (shows QR code)
notifybridge pair

# Send a test notification
notifybridge send "Hello from Mac!"
```

## Installation

### From source

```bash
git clone https://github.com/felipersas/notifybridge.git
cd notifybridge
make build
make install   # copies to /usr/local/bin
```

### With Go

```bash
go install github.com/felipersas/notifybridge/cmd/notifybridge@latest
```

## Requirements

- **macOS** with SSH client
- **Android** with [Termux](https://termux.dev/) + [Termux:API](https://wiki.termux.com/wiki/Termux:API)
- **Tailscale** on both devices (recommended) or same network

## Configuration

### `~/.notifybridge.conf`

Main configuration file (created by `notifybridge pair` or `notifybridge setup`):

```bash
# Tailscale IP of the Android device
ANDROID_IP="100.115.83.120"

# SSH user on Termux
SSH_USER="root"

# SSH port (Termux sshd default)
SSH_PORT="8022"

# Default notification title
DEFAULT_TITLE="Mac Remote"

# Sound on notification
SOUND=true

# Notification priority (high|default|low|max|min)
PRIORITY="high"

# Max retry attempts
MAX_RETRIES=2

# Retry delay in seconds
RETRY_DELAY=3
```

Set `NOTIFYBRIDGE_CONF` to override config file path.

### `~/.notifybridge-projects.conf`

Per-project notification styling:

```bash
# Format: DIR_NAME=EMOJI:LED_COLOR:PRIORITY
# LED_COLOR: hex without # (FF0000=red, 4CAF50=green, etc)
# PRIORITY: high|default|low|max|min

# Default (fallback)
DEFAULT=⚡:2196F3:high

# Projects
Termux=🔧:2196F3:high
licespot-web=🛒:4CAF50:high
licespot-api=🛒:4CAF50:high
AWS=☁️:FF9800:default
```

Set `NOTIFYBRIDGE_PROJECTS` to override projects file path.

## Claude Code Hook Setup

Add to `.claude/hooks.json` in your project:

```json
{
  "hooks": {
    "Stop": [{
      "type": "command",
      "command": "notifybridge hook"
    }]
  }
}
```

When Claude Code finishes a task, you get a notification on your phone with:
- Project-specific emoji and LED color
- Last assistant message (truncated to 200 chars)
- Grouped by project directory

## Commands

| Command | Description |
|---------|-------------|
| `notifybridge send MSG` | Send a notification |
| `notifybridge hook` | Claude Code hook (reads JSON from stdin) |
| `notifybridge pair` | Pair with Android via QR code |
| `notifybridge setup` | Interactive setup wizard |
| `notifybridge test` | Test connectivity to Android |
| `notifybridge config show` | Show current configuration |
| `notifybridge version` | Print version |

### `send` flags

```
-t, --title string     Notification title
    --led-color string LED color hex (e.g. FF0000)
    --group string     Notification group key
    --id string        Notification channel ID
    --priority string  high|default|low|max|min
    --sound            Play notification sound (default true)
```

## Pairing Flow

1. Run `notifybridge pair` on Mac — shows a QR code
2. Scan QR on phone (or open URL in Termux browser)
3. Run the displayed `curl ... | bash` command in Termux
4. Termux installs openssh, termux-api, generates SSH keys, starts sshd
5. SSH keys are exchanged bidirectionally
6. `claude-remote` command is installed on phone for reverse access
7. Config is written automatically on Mac

After pairing, use `claude-remote` on your phone to SSH into a Claude Code tmux session on your Mac.

## Architecture

```
Claude Code Hook ──> hook.Run() ──> profile.Match() + notify.SSHNotifier
                                                   │
                                                   ▼
                                            SSH to Termux
                                                   │
                                                   ▼
                                         termux-notification
```

| Package | Purpose |
|---------|---------|
| `cmd/notifybridge` | CLI entry point (cobra) |
| `internal/cfg` | Configuration loading and saving |
| `internal/profile` | Project profile matching |
| `internal/notify` | SSH notification delivery with retry |
| `internal/hook` | Claude Code Stop hook handler |
| `internal/pair` | QR code pairing and SSH key exchange |
| `internal/setup` | Interactive setup wizard |
| `internal/health` | Connectivity testing |

## Manual Setup (Alternative)

If you prefer not to use the pairing wizard:

1. **Android (Termux):** Run `scripts/termux-setup.sh`
2. **macOS:** Run `scripts/mac-setup.sh`

## License

MIT — see [LICENSE](LICENSE)
