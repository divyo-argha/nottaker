# nottaker

> Lightweight multi-tab auto-saving scratchpad — TUI + GUI, built in Go.

```
✦ nottaker
 1: scratch   2: ideas   3: todo   [+]
╭──────────────────────────────────────────╮
│ Start typing…                             │
│                                           │
│                                           │
╰──────────────────────────────────────────╯
^N new  ^W close  ^→/← switch  Tab cycle  ^C quit          ✓ saved 21:04:55
```

## Features

- **Multi-tab workspace** — create, destroy, and cycle through tabs instantly
- **Auto-save on every keystroke** — no Ctrl+S, no lost work, ever
- **Crash-safe persistence** — atomic temp-file + rename writes
- **Platform-aware storage** — `~/Library/Application Support/nottaker/` on macOS, `~/.config/nottaker/` on Linux, `%AppData%/nottaker/` on Windows
- **Two interfaces, one backend** — TUI (Bubble Tea) and GUI (Wails) share identical Go persistence code

---

## Project Structure

```
nottaker/
├── core/storage.go        # Shared persistence layer (Tab, State, async save)
├── tui/main.go            # Bubble Tea TUI application
├── gui/
│   ├── main.go            # Wails entry point
│   ├── app.go             # JS-bound Go methods
│   ├── assets.go          # Embed directive for frontend
│   ├── wails.json         # Wails project config
│   └── frontend/
│       ├── index.html
│       ├── style.css
│       └── main.js
├── npm/
│   ├── package.json       # npm distribution config
│   ├── bin/nottaker.js    # CLI shim for npx
│   └── scripts/install.js # Platform binary downloader
└── Makefile
```

---

## Quick Start

### Prerequisites

- Go 1.22+
- [Wails v2](https://wails.io/docs/gettingstarted/installation) (for GUI only)

### TUI

```bash
# Build
make tui

# Run
./nottaker

# Or run directly
go run ./tui/
```

### GUI Desktop App

```bash
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Build
make gui
# Output: gui/build/bin/nottaker-gui.app (macOS)

# Development mode (hot reload)
make dev-gui
```

### via npx (TUI only)

```bash
npx nottaker
```

> On first run, `postinstall` downloads the pre-compiled binary for your platform.
> Requires internet access. Subsequent runs use the cached binary.

---

## Keyboard Shortcuts (TUI & GUI)

| Shortcut | Action |
|---|---|
| `Ctrl+N` | New tab |
| `Ctrl+W` | Close current tab |
| `Ctrl+Tab` | Next tab |
| `Ctrl+Shift+Tab` | Previous tab |
| `Ctrl+1…9` | Jump to tab by number |
| `Ctrl+C` *(TUI)* | Quit |

### GUI Only
| Action | How |
|---|---|
| Rename tab | Double-click tab title, Enter to confirm, Esc to cancel |

---

## Persistence

State is stored as `state.json` in your platform config directory:

| Platform | Path |
|---|---|
| macOS | `~/Library/Application Support/nottaker/state.json` |
| Linux | `~/.config/nottaker/state.json` |
| Windows | `%APPDATA%\nottaker\state.json` |

Writes are **atomic** — a `.tmp` file is written first, then renamed, so a crash mid-write never corrupts your data.

---

## Publishing to npm

```bash
# Cross-compile all platform binaries
make cross

# Publish npm package (binaries must be uploaded to GitHub Releases first)
cd npm && npm publish
```

---

## License

MIT
