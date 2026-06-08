# octoNote

> **A lightning-fast, crash-proof, multi-tab scratchpad for your terminal and desktop.**

**octoNote** is a radically simple text workspace designed for ephemeral notes, scratchpad ideas, and quick to-dos. It offers an identical experience whether you are in the terminal (TUI) or using the desktop application (GUI). Built completely in Go, octoNote completely eliminates the concept of "saving". Every keystroke is instantly and safely committed to disk using an atomic file-write process, meaning a crash or accidental close never loses your work.

It works completely across all platforms out of the box (macOS, Linux, and Windows) for both the Terminal UI and the Desktop GUI.

```
✦ octonote
 1: scratch   2: ideas   3: todo   [+]
╭──────────────────────────────────────────╮
│ Start typing…                            │
│                                          │
│                                          │
╰──────────────────────────────────────────╯
^N new  ^W close  ^→/← switch  Tab cycle  ^C quit          ✓ saved 21:04:55
```

## Features

- **Multi-tab workspace** — create, destroy, and cycle through tabs instantly
- **Auto-save on every keystroke** — no Ctrl+S, no lost work, ever
- **Crash-safe persistence** — atomic temp-file + rename writes
- **Platform-aware storage** — `~/Library/Application Support/octonote/` on macOS, `~/.config/octonote/` on Linux, `%AppData%/octonote/` on Windows
- **Two interfaces, one backend** — TUI (Bubble Tea) and GUI (Wails) share identical Go persistence code

---

## Project Structure

```
octonote/
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
│   ├── bin/octonote.js    # CLI shim for npx
│   └── scripts/install.js # Platform binary downloader
└── Makefile
```

---

## Installation

### One-Line Installers

The easiest way to get `octoNote` is to use our installation scripts, which download the pre-compiled binary for your system.

**macOS / Linux (Bash)**
```bash
# Install TUI + GUI
curl -sSfL https://raw.githubusercontent.com/divyo-argha/octonote/main/scripts/install-gui.sh | sh

# Install TUI only
curl -sSfL https://raw.githubusercontent.com/divyo-argha/octonote/main/scripts/install-cli.sh | sh
```

**Windows (PowerShell)**
```powershell
# Install TUI + GUI
iwr -useb https://raw.githubusercontent.com/divyo-argha/octonote/main/scripts/install-gui.ps1 | iex

# Install TUI only
iwr -useb https://raw.githubusercontent.com/divyo-argha/octonote/main/scripts/install-cli.ps1 | iex
```

### NPM / NPX (TUI Only)

For Node.js users, the TUI is packaged as an npm executable. When installed, a `postinstall` script automatically fetches the correct pre-compiled Go binary for your exact OS and architecture.

```bash
# Install globally
npm install -g octonote

# Run from anywhere
octonote
```

Alternatively, run without installing:
```bash
npx octonote
```

### Build from Source

**Prerequisites**
- Go 1.22+
- [Wails v2](https://wails.io/docs/gettingstarted/installation) (for GUI only)

```bash
# Build the TUI
make tui
./octonote

# Build the GUI Desktop App
# Output goes to gui/build/bin/
make gui
```

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

## Persistence

State is stored as `state.json` in your platform config directory:

| Platform | Path |
|---|---|
| macOS | `~/Library/Application Support/octonote/state.json` |
| Linux | `~/.config/octonote/state.json` |
| Windows | `%APPDATA%\octonote\state.json` |

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
