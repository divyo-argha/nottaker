<div align="center">
  <br />
  <img src="assets/logo.png" alt="octoNote" width="120" height="120" style="border-radius:26px" />
  <!-- <br /><br /> -->
  <h1>octoNote</h1>

  <p>
    <strong>A lightning-fast, crash-proof, multi-tab scratchpad for terminal and desktop.</strong><br />
    Stop hitting Ctrl+S. Stop losing quick ideas. Stop naming throwaway files.
  </p>

  <p>
    <a href="https://github.com/divyo-argha/octonote/releases"><img src="https://img.shields.io/github/v/release/divyo-argha/octonote?style=flat-square&color=00FFAA&label=latest" alt="Latest Release" /></a>
    <a href="https://github.com/divyo-argha/octonote/releases"><img src="https://img.shields.io/github/downloads/divyo-argha/octonote/total?style=flat-square&color=00FFAA&label=gh%20downloads" alt="GitHub Downloads" /></a>
    <a href="https://www.npmjs.com/package/octonote"><img src="https://img.shields.io/npm/v/octonote?style=flat-square&color=CB3837&logo=npm&logoColor=white&label=npm" alt="npm" /></a>
    <a href="https://pkg.go.dev/github.com/divyo-argha/octonote"><img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" /></a>
    <a href="LICENSE"><img src="https://img.shields.io/badge/License-Apache_2.0-blue?style=flat-square" alt="Apache 2.0" /></a>
  </p>

  <p>
    <a href="#-the-problem">The Problem</a> ·
    <a href="#-install">Install</a> ·
    <a href="#-quick-start">Quick Start</a> ·
    <a href="#-why-octonote">Why octoNote</a> ·
    <a href="#-features">Features</a> ·
    <a href="#-keyboard-shortcuts">Shortcuts</a> ·
    <a href="#-p2p-sharing">Sharing</a> ·
    <a href="#-persistence">Persistence</a> ·
    <a href="#-contributing">Contributing</a>
  </p>

  <br />

  <img src="https://img.shields.io/badge/macOS-supported-000000?style=for-the-badge&logo=apple&logoColor=white" alt="macOS" />
  <img src="https://img.shields.io/badge/Linux-supported-FCC624?style=for-the-badge&logo=linux&logoColor=black" alt="Linux" />
  <img src="https://img.shields.io/badge/Windows-supported-0078D4?style=for-the-badge&logo=windows&logoColor=white" alt="Windows" />
  <img src="https://img.shields.io/badge/Terminal_TUI-supported-4E9A06?style=for-the-badge&logo=gnubash&logoColor=white" alt="Terminal TUI" />
  <img src="https://img.shields.io/badge/Desktop_GUI-supported-0052CC?style=for-the-badge&logo=wails&logoColor=white" alt="Desktop GUI" />

  <br /><br />

</div>

---

## 😤 The Problem

You're a developer coding away. You need to quickly copy a JSON payload, sketch a rough database schema, draft an email, or keep a temporary todo checklist.

And this happens every day:

```
# You open VS Code or Notepad.
# You type your notes.
# Later, you want to close it:
"Do you want to save changes to Untitled-1?"  ← 💀 annoying prompt.

# Or you name it temp_123.txt and throw it in ~/Desktop.
# Now your Desktop is a graveyard of 80 unnamed text files.
# Or your computer restarts for an update and your unsaved buffer is gone forever.
```

You've tried everything:

| Attempt | Result |
|---------|--------|
| Keeping unsaved VS Code tabs | Heavy resource usage, clutters workspace |
| Standard Apple Notes / Keep | Slow to open, requires mouse navigation |
| Creating `temp.txt` everywhere | litters your codebase repositories |
| Terminal scratchpads (`nano /tmp/t`) | Single tab, manual save, wiped on reboot |

**octoNote is the permanent fix.** A dedicated scratchpad workspace that opens instantly, supports multiple tabs, and auto-saves every single character to disk. You never name a scratch file, and you never hit save.

---

## 📦 Install

<table>
<tr>
<td width="33%" valign="top">

### macOS / Linux (Bash)
```bash
# Install TUI & GUI in one command
curl -sSfL https://raw.githubusercontent.com/divyo-argha/octonote/main/scripts/install.sh | sh
```

</td>
<td width="33%" valign="top">

### Windows (PowerShell)
```powershell
# Install TUI & GUI in one command
iwr -useb https://raw.githubusercontent.com/divyo-argha/octonote/main/scripts/install.ps1 | iex
```

</td>
<td width="33%" valign="top">

### NPM
```bash
# Install TUI & GUI globally in one command
npm install -g octonote

# Run TUI instantly without install
npx octonote
```

</td>
</tr>
</table>

**Requirements:** macOS, Linux, or Windows. Go 1.22+ & Wails v2 (only if building from source).

---

## ⚡ Quick Start

Type `octonote` in your terminal to start the Bubble Tea TUI, or launch the Desktop App to use the GUI.

```bash
# Launch the terminal scratchpad
octonote
```

```
✦ octonote
 1: scratch   2: ideas   3: todo   [+]
╭──────────────────────────────────────────╮
│ Start typing…                            │
│                                          │
╰──────────────────────────────────────────╯
^N new  ^W close  ^→/← switch  Tab cycle  ^C quit          ✓ saved 21:04:55
```

Write notes, open multiple tabs, close it, and reopen it. Everything is exactly where you left it.

---

## 🏆 Why octoNote?

| Feature | octoNote | VS Code / IDE | standard Notes App | `nano` / `vim` |
|---------|:--------:|:-------------:|:------------------:|:--------------:|
| Instantly open | ✅ | ❌ | ⚠️ slow | ✅ |
| Keyboard-driven TUI | ✅ | ✅ | ❌ | ✅ |
| Multi-tab workspace | ✅ | ✅ | ❌ | ❌ |
| Auto-save on every keystroke | ✅ | ⚠️ auto-save delay | ⚠️ sync delay | ❌ |
| Crash-proof atomic writes | ✅ | ❌ | ❌ | ❌ |
| Identical Terminal & Desktop states | ✅ | ❌ | ❌ | ❌ |
| Zero file naming overhead | ✅ | ❌ | ❌ | ❌ |
| P2P tab sharing (no account) | ✅ | ❌ | ❌ | ❌ |

---

## ✨ Features

<table>
<tr>
<td width="50%" valign="top">

### 📝 Zero-Management Notes
- Create and destroy tabs on the fly with shortcuts.
- No "Save as..." prompts.
- No files or directory trees to organize.
- Lightweight, low-memory footprint.

</td>
<td width="50%" valign="top">

### 🛡️ Crash-Proof Engine
- Saves state on every single keystroke.
- **Atomic file writes**: Writes to a `.tmp` file and swaps it instantly.
- Crash, power loss, or terminal termination will never corrupt or lose data.

</td>
</tr>
<tr>
<td width="50%" valign="top">

### 💻 Dual Interfaces, Unified State
- Terminal TUI built with **Bubble Tea** (Lip Gloss, bubbletier).
- Desktop GUI built with **Wails** (Vanilla HTML/CSS/JS frontend).
- Both interfaces interact with the identical Go storage core.
- Share your notes between CLI and Desktop seamlessly.

</td>
<td width="50%" valign="top">

### 🚀 Zero Configuration
- Works out of the box on Windows, macOS, and Linux.
- Data stored in proper platform-specific config directories.
- Small binary size, rapid startup.

</td>
</tr>
<tr>
<td width="50%" valign="top">

### 🔗 P2P Tab Sharing
- Share any tab instantly with a short human-readable code.
- **No account, no cloud, no server setup** required.
- End-to-end encrypted via [SPAKE2](https://www.ietf.org/rfc/rfc9382.html) — relay never sees your content.
- Codes are one-time use and expire immediately after transfer.
- Works between TUI and GUI, and is compatible with the Python `magic-wormhole` client.

</td>
<td width="50%" valign="top">

### 🚀 Zero Configuration
- Works out of the box on Windows, macOS, and Linux.
- Data stored in proper platform-specific config directories.
- Small binary size, rapid startup.

</td>
</tr>
</table>

---

## ⌨️ Keyboard Shortcuts

Both TUI and GUI share the same quick, muscle-memory shortcuts:

| Shortcut | Action |
|---|---|
| `Ctrl+N` | Create new tab |
| `Ctrl+W` | Delete / close current tab |
| `Ctrl+Tab` | Switch to next tab |
| `Ctrl+Shift+Tab` | Switch to previous tab |
| `Ctrl+1` ... `Ctrl+9` | Jump directly to tab by index |
| `Tab` *(TUI)* | Cycle focus between tab bar and text area |
| `Ctrl+S` *(TUI)* | Share active tab — generates a wormhole code |
| `Ctrl+R` *(TUI)* | Receive mode — enter a peer's code to import their tab |
| `Ctrl+C` *(TUI)* | Quit application (or cancel active share) |

---

## 🔗 P2P Sharing

octoNote has built-in **serverless peer-to-peer tab sharing** powered by the [Magic Wormhole](https://magic-wormhole.readthedocs.io/) protocol.

**In the GUI:** Click the **Share** button in the title bar.  
**In the TUI:** Press `Ctrl+S` to share, `Ctrl+R` to receive.

```
# Sender (TUI example)
share code:  7-crossover-alpha   waiting for peer…   ^C cancel

# Receiver (TUI example)
enter code: _7-crossover-alpha_  then ↵ to connect  Esc cancel
```

- No account needed.
- Codes are one-time use and expire instantly after the transfer.
- Content is end-to-end encrypted — the relay server never sees your notes.
- Interoperable with the Python `magic-wormhole` reference client.

📖 **Full documentation:** [SHARING.md](SHARING.md)

---

## 📁 Persistence

Notes are saved as `state.json` inside your platform's standard configuration directory:

| Platform | Path |
|---|---|
| **macOS** | `~/Library/Application Support/octonote/state.json` |
| **Linux** | `~/.config/octonote/state.json` |
| **Windows** | `%APPDATA%\octonote\state.json` |

---

## 🛠️ Build from Source

If you want to build `octoNote` locally, make sure you have Go 1.22+ and Wails v2 installed:

```bash
# Clone the repository
git clone https://github.com/divyo-argha/octonote.git
cd octonote

# Build TUI (outputs binary to the root directory)
make tui

# Build GUI Desktop App (outputs to gui/build/bin/)
make gui
```

---

## 📄 License

Apache License, Version 2.0 — see [LICENSE](LICENSE) for details.

---

<div align="center">

**Ephemeral notes, permanent safety.**

<br />

[![GitHub](https://img.shields.io/badge/Star%20on%20GitHub-181717?style=for-the-badge&logo=github&logoColor=white)](https://github.com/divyo-argha/octonote)
[![npm](https://img.shields.io/badge/Install%20via%20npm-CB3837?style=for-the-badge&logo=npm&logoColor=white)](https://www.npmjs.com/package/octonote)

<br />

<sub>If octoNote saved you from losing your scratch notes, consider giving it a ⭐</sub>

</div>
