# TAVPBox

> LXC-based dev environment — like Lando, but lighter RAM.

## Install

### Windows (PowerShell as Administrator)
```powershell
iex (irm 'https://get.tavp.dev/setup-tavpbox.ps1' -UseB)
```

### macOS
```bash
curl -fsSL https://get.tavp.dev/setup-tavpbox.sh | bash
```

### Linux
```bash
sudo curl -fsSL https://get.tavp.dev/setup-tavpbox.sh | bash
```

## Quick Start

```bash
tavpbox init      # First-time setup wizard
tavpbox create    # Create your first box
tavpbox list      # List all boxes
```

## Commands

| Command | Description |
|---------|-------------|
| `tavpbox init` | First-time setup wizard |
| `tavpbox create` | Create a new box (TUI or from file) |
| `tavpbox start <name>` | Start a box |
| `tavpbox stop <name>` | Stop a box |
| `tavpbox list` | List all boxes |
| `tavpbox ssh <name>` | Enter box terminal |
| `tavpbox info <name>` | Show box info |
| `tavpbox destroy <name>` | Destroy a box |
| `tavpbox rebuild <name>` | Recreate box (data preserved) |
| `tavpbox snapshot <name>` | Create a snapshot |
| `tavpbox status` | Show system status |
| `tavpbox logs <name>` | Display logs |
| `tavpbox exec <name> <cmd>` | Execute command |

## Config

### `.tavpbox.yml` (per-project)
```yaml
name: my-project
stack: tavp
services:
  - mariadb
  - redis
  - mailpit
webroot: .
```

### `~/.tavpbox/config.yml` (global)
```yaml
domain_suffix: tavp.local
default_ram: 512MB
default_cpu: 1
default_distro: ubuntu/24.04
```

## RAM Comparison

```
20 projects running:
  Docker/Lando:  ~3.2GB
  TAVPBox:       ~745MB (75% less!)
```

## License

MIT
