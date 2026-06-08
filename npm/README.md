<div align="center">
  <br />
  <img src="https://raw.githubusercontent.com/divyo-argha/octonote/main/assets/logo.png" alt="octoNote" width="120" height="120" style="border-radius:26px" />
  <!-- <br /><br /> -->
  <h1>octoNote</h1>

  <p>
    <strong>A lightning-fast, crash-proof, multi-tab scratchpad for your terminal.</strong><br />
    Stop hitting Ctrl+S. Stop losing quick ideas. Stop naming throwaway files.
  </p>

  <p>
    <a href="https://github.com/divyo-argha/octonote/releases"><img src="https://img.shields.io/github/v/release/divyo-argha/octonote?style=flat-square&color=00FFAA&label=latest" alt="Latest Release" /></a>
    <a href="https://www.npmjs.com/package/octonote"><img src="https://img.shields.io/npm/v/octonote?style=flat-square&color=CB3837&logo=npm&logoColor=white&label=npm" alt="npm" /></a>
    <a href="https://www.npmjs.com/package/octonote"><img src="https://img.shields.io/npm/dt/octonote?style=flat-square&color=CB3837&logo=npm&logoColor=white&label=total%20downloads" alt="total downloads" /></a>
    <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-22c55e?style=flat-square" alt="MIT" /></a>
  </p>

  <p>
    <a href="#-the-problem">The Problem</a> ·
    <a href="#-installation">Installation</a> ·
    <a href="#-quick-start">Quick Start</a> ·
    <a href="#-why-octonote">Why octoNote</a> ·
    <a href="#-features">Features</a> ·
    <a href="#-keyboard-shortcuts">Shortcuts</a> ·
    <a href="#-persistence">Persistence</a>
  </p>

  <br />

  <img src="https://img.shields.io/badge/macOS-supported-000000?style=for-the-badge&logo=apple&logoColor=white" alt="macOS" />
  <img src="https://img.shields.io/badge/Linux-supported-FCC624?style=for-the-badge&logo=linux&logoColor=black" alt="Linux" />
  <img src="https://img.shields.io/badge/Windows-supported-0078D4?style=for-the-badge&logo=windows&logoColor=white" alt="Windows" />
  <img src="https://img.shields.io/badge/Terminal_TUI-supported-4E9A06?style=for-the-badge&logo=gnubash&logoColor=white" alt="Terminal TUI" />

  <br /><br />

</div>

---

> [!NOTE]
> This npm package distributes the **Terminal User Interface (TUI)** version of octoNote. If you want the Desktop GUI version, please check the main [GitHub Repository](https://github.com/divyo-argha/octonote).

---

## 😤 The Problem

You're a developer coding away in your terminal. You need to quickly copy a JSON payload, sketch a rough database schema, draft an email, or keep a temporary todo checklist.

And this happens every day:

```
# You open nano or vim.
# You type your notes.
# Later, you want to close it:
"Save modified buffer?"  ← 💀 annoying prompt.

# Or you name it temp_123.txt and throw it in ~/Desktop.
# Now your Desktop is a graveyard of 80 unnamed text files.
# Or your terminal session closes, and your unsaved buffer is gone forever.
```

**octoNote is the permanent fix.** A dedicated scratchpad workspace that opens instantly inside your terminal, supports multiple tabs, and auto-saves every single character to disk. You never name a scratch file, and you never hit save.

---

## 📦 Installation

This npm package automatically downloads the optimized Go binary matching your operating system and architecture upon installation.

```bash
# Install globally
npm install -g octonote

# Or run instantly without installation using npx
npx octonote
```

---

## ⚡ Quick Start

Launch the terminal scratchpad with one command:

```bash
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

Write notes, open multiple tabs, exit anytime, and reopen. Everything is exactly where you left it.

---

## 🏆 Why octoNote?

| Feature | octoNote | `nano` / `vim` | standard Notes App |
|---------|:--------:|:--------------:|:------------------:|
| Instantly open | ✅ | ✅ | ❌ |
| Keyboard-driven TUI | ✅ | ✅ | ❌ |
| Multi-tab workspace | ✅ | ❌ | ❌ |
| Auto-save on every keystroke | ✅ | ❌ | ⚠️ sync delay |
| Crash-proof atomic writes | ✅ | ❌ | ❌ |
| Zero file naming overhead | ✅ | ❌ | ❌ |

---

## ✨ Features

- 📝 **Zero-Management Notes** — Create and destroy tabs on the fly with shortcuts. No "Save as..." prompts. No files or directory trees to organize.
- 🛡️ **Crash-Proof Engine** — Saves state on every single keystroke. Uses atomic temp-file swaps so you never lose or corrupt your notes.
- 💻 **Platform-Native Storage** — Saves to standardized locations depending on your OS.
- 🚀 **Built in Go** — Ultra-lightweight, extremely fast startup, minimal memory consumption.

---

## ⌨️ Keyboard Shortcuts

| Shortcut | Action |
|---|---|
| `Ctrl+N` | Create new tab |
| `Ctrl+W` | Delete / close current tab |
| `Ctrl+Tab` | Switch to next tab |
| `Ctrl+Shift+Tab` | Switch to previous tab |
| `Ctrl+1` ... `Ctrl+9` | Jump directly to tab by index |
| `Tab` | Cycle focus between tab bar and text area |
| `Ctrl+C` | Quit application |

---

## 📁 Persistence

Notes are saved as `state.json` inside your platform's standard configuration directory:

- **macOS:** `~/Library/Application Support/octonote/state.json`
- **Linux:** `~/.config/octonote/state.json`
- **Windows:** `%APPDATA%\octonote\state.json`

---

## 📄 License

MIT — see [LICENSE](https://github.com/divyo-argha/octonote/blob/main/LICENSE) for details.
