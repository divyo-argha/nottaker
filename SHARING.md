# 🔗 P2P Tab Sharing

> Share any scratchpad tab with a peer in seconds — no account, no cloud storage, no setup.
> Just a short code like **`7-crossover-alpha`**.

---

## Table of Contents

- [How It Works](#how-it-works)
- [Security Model](#security-model)
- [User Guide](#user-guide)
  - [GUI — Desktop App](#gui--desktop-app)
  - [TUI — Terminal](#tui--terminal)
- [Cross-Compatibility](#cross-compatibility)
- [Troubleshooting](#troubleshooting)
- [Architecture](#architecture)
  - [Protocol Overview](#protocol-overview)
  - [Wire Format](#wire-format)
  - [Code Reference](#code-reference)
- [Privacy FAQ](#privacy-faq)

---

## How It Works

octoNote uses the [Magic Wormhole](https://magic-wormhole.readthedocs.io/) protocol, specifically the Go implementation [`wormhole-william`](https://github.com/psanford/wormhole-william).

```
User A (sender)              Public Relay              User B (receiver)
      │                          │                           │
      │── Register code ────────>│                           │
      │   "7-crossover-alpha"    │                           │
      │                          │<─── Connect via code ─────│
      │                          │                           │
      │<═══════════════ SPAKE2 cryptographic handshake ═════>│
      │         (relay authenticates but cannot read)        │
      │                          │                           │
      │══════════ Tab content streams directly P2P ══════════>│
      │            (relay is now out of the picture)         │
```

**The relay server acts as a matchmaker only.** It facilitates the handshake but never sees your note content — all transferred data is end-to-end encrypted using [SPAKE2](https://www.ietf.org/rfc/rfc9382.html), a Password-Authenticated Key Exchange algorithm purpose-built for deriving strong keys from short, human-readable codes.

> [!NOTE]
> The public relay used by default is `relay.magic-wormhole.io`, maintained by the open-source Magic Wormhole project. No registration is required to use it.

---

## Security Model

| Property | Details |
|---|---|
| **Encryption** | SPAKE2 end-to-end — relay server cannot decrypt content |
| **Authentication** | The code *is* the password — both sides must know it |
| **One-time use** | Each code can only be used once; it expires immediately after transfer |
| **No persistence** | Neither the relay nor octoNote stores your note on any server |
| **Forward secrecy** | Each session generates a fresh ephemeral key pair |

> [!IMPORTANT]
> Treat the share code like a **temporary password**. Anyone who learns the code before your peer connects can intercept the transfer. Share it via a trusted channel (e.g., a private chat message).

> [!TIP]
> Codes expire the moment the transfer completes or either side cancels. There is no need to "revoke" a used code.

---

## User Guide

### GUI — Desktop App

#### Sending a tab (Host)

1. Open the **octoNote** desktop app.
2. Navigate to the tab you want to share.
3. Click the **Share** button in the top-left title bar area (next to the logo).
4. In the modal that opens, make sure you're on the **📤 Share (Host)** tab.
5. Click **Generate Code & Share**.
   - octoNote connects to the relay (~1 second) and displays your code:

     ```
     7-crossover-alpha
     ```

6. Send this code to your peer via any channel (chat, email, verbal).
7. The modal shows **⏳ Waiting for peer to connect…** — stay on this screen.
8. Once your peer enters the code, the transfer happens instantly.
9. The modal shows **✅ Tab sent successfully!** and closes automatically.

#### Receiving a tab (Guest)

1. Click the **Share** button.
2. Switch to the **📥 Receive (Guest)** tab.
3. Paste or type the code from your peer (e.g. `7-crossover-alpha`).
4. Click **Connect**.
5. The modal shows **⏳ Connecting to peer…**
6. Once the transfer completes, the modal shows:
   ```
   ✅ Tab "my-tab-title" added to your notebook!
   ```
   The new tab appears automatically and becomes the active tab.

> [!TIP]
> You can click the **copy** button next to the code to copy it to your clipboard in one click.

---

### TUI — Terminal

The TUI implements the full share/receive flow inline — no separate window is opened. Status appears in the **legend bar** at the bottom of the screen.

#### Sending a tab (Host)

1. Navigate to the tab you want to share.
2. Press **`Ctrl+S`**.
3. The legend bar changes to:
   ```
   opening wormhole…
   ```
   then after ~1 second:
   ```
   share code:  7-crossover-alpha   waiting for peer…   ^C cancel
   ```
4. Read the code to your peer.
5. Once the peer connects, the legend bar returns to normal — transfer is complete.
6. Press **`Ctrl+C`** at any point to cancel a pending share (this does **not** quit the app while a share is active; a second `Ctrl+C` will quit).

#### Receiving a tab (Guest)

1. Press **`Ctrl+R`**.
2. The legend bar shows:
   ```
   enter code: _  _ then ↵ to connect  Esc cancel
   ```
3. Type the code character by character. Use `Backspace` to correct mistakes.
4. Press **`Enter`** to connect.
5. The bar shows a connecting state, then the new tab appears instantly in your tab bar and becomes active.
6. Press **`Esc`** or **`Ctrl+C`** to cancel.

**Updated keyboard shortcut reference for TUI:**

| Shortcut | Action |
|---|---|
| `Ctrl+S` | Share the active tab (host/sender role) |
| `Ctrl+R` | Enter receive mode (guest/receiver role) |
| `Ctrl+C` *(while sharing)* | Cancel the pending share |
| `Ctrl+C` *(normal mode)* | Quit |

---

## Cross-Compatibility

octoNote's share format is interoperable with the reference [magic-wormhole](https://pypi.org/project/magic-wormhole/) Python client and any other compliant implementation.

| Scenario | Works? | Notes |
|---|---|---|
| octoNote GUI → octoNote GUI | ✅ | Full fidelity — title preserved |
| octoNote TUI → octoNote TUI | ✅ | Full fidelity — title preserved |
| octoNote GUI → octoNote TUI | ✅ | Full fidelity — title preserved |
| octoNote → `wormhole receive` (Python) | ✅ | Receives raw JSON payload |
| `wormhole send` (Python) → octoNote | ✅ | Treated as plain text, titled "shared note" |
| octoNote → `wormhole-william` CLI | ✅ | Receives raw JSON payload |

> [!NOTE]
> When receiving from a non-octoNote sender (e.g. Python `magic-wormhole`), the tab is created with the title **"shared note"** and the raw text content as the body.

---

## Troubleshooting

### "share: receive handshake" error

The code was wrong, already used, or has expired. Ask the sender to generate a new code.

### "connection closed before transfer" error

The sender cancelled or closed octoNote before you connected. Ask them to share again.

### The code never appears / "opening wormhole…" hangs

- Check your internet connection.
- A firewall may be blocking WebSocket connections to `relay.magic-wormhole.io` on port `4000`. Try from a different network or ask your sysadmin to allow it.

### "peer sent a file, not text" error

The sender is using the reference `magic-wormhole` CLI with `wormhole send <file>` instead of `wormhole send --text`. Only text transfers are supported by octoNote.

### Share modal opens but "Generate Code & Share" does nothing

The Wails Go bridge may not have initialized yet. Wait a moment and try again, or restart the GUI.

---

## Architecture

### Protocol Overview

The sharing feature is implemented in three layers:

```
┌─────────────────────────────────────────────────────────────────────┐
│  UI Layer                                                           │
│  ┌──────────────────────────┐   ┌──────────────────────────────┐   │
│  │  TUI (Bubble Tea)        │   │  GUI (Wails + HTML/CSS/JS)   │   │
│  │  Ctrl+S / Ctrl+R         │   │  Share button + Modal        │   │
│  │  Legend bar status       │   │  Wails events (share:*)      │   │
│  └────────────┬─────────────┘   └──────────────┬───────────────┘   │
│               │                                │                   │
│               └───────────────┬────────────────┘                   │
│                               ▼                                     │
├─────────────────────────────────────────────────────────────────────┤
│  Core Layer  (core/share.go)                                        │
│                                                                     │
│  ShareSend(ctx, tab) → (code, waitFn, err)                         │
│  ShareReceive(ctx, code) → (ShareResult, err)                      │
│                                                                     │
│  Serialises tabs as JSON, delegates to wormhole-william             │
├─────────────────────────────────────────────────────────────────────┤
│  Transport Layer  (github.com/psanford/wormhole-william)            │
│                                                                     │
│  SPAKE2 handshake  ·  WebSocket relay  ·  Encrypted data channel   │
└─────────────────────────────────────────────────────────────────────┘
```

### Wire Format

octoNote sends a compact JSON envelope as the wormhole text payload:

```json
{
  "tab_title": "my api notes",
  "body": "POST /users\nContent-Type: application/json\n...",
  "sender_label": "Alice",
  "version": 1
}
```

| Field | Type | Description |
|---|---|---|
| `tab_title` | string | The tab's display title as shown in the tab bar |
| `body` | string | Full text content of the tab, newlines preserved |
| `sender_label` | string | (Optional) The name/label of the sender, e.g. "Alice" |
| `version` | int | Schema version — always `1` in the current implementation |

If the received payload cannot be parsed as this JSON schema (e.g. it came from the Python CLI), the raw text is used as the body and the title defaults to `"shared note"`.

### Code Reference

#### `core/share.go`

The single source of truth for the share protocol. Both the TUI and GUI import from here.

```go
ShareSend(ctx context.Context, tab Tab, senderLabel string) (code string, wait func() error, err error)
```
- Serialises `tab` into a `SharePayload` JSON string with the `senderLabel` included.
- Calls `wormhole.Client.SendText()` to register with the relay and obtain a code.
- Returns the code immediately so the UI can display it to the user.
- Returns a `wait` function — the caller **must** call this on a goroutine to block until the peer receives the data or until `ctx` is cancelled.

```go
ShareReceive(ctx context.Context, code string) (ShareResult, error)
```
- Calls `wormhole.Client.Receive()` to perform the SPAKE2 handshake.
- Validates that the transfer type is `TransferText` (rejects files).
- Reads the full payload via `io.ReadAll`.
- Unmarshals the JSON envelope; falls back to raw text if the sender is not octoNote.

#### `gui/app.go` — Wails bridge methods

| Method | JS call | Description |
|---|---|---|
| `ShareSend()` | `go.main.App.ShareSend()` | Starts a share on a goroutine; emits `share:code` then `share:done` or `share:error` |
| `ShareReceive(code)` | `go.main.App.ShareReceive(code)` | Connects as guest; emits `share:received` or `share:error` |
| `ShareCancel()` | `go.main.App.ShareCancel()` | Cancels any in-flight share or receive via context cancellation |

**Wails events emitted by Go → consumed by `main.js`:**

| Event | Payload | When |
|---|---|---|
| `share:code` | `string` — the wormhole code | Relay registered, code ready to share |
| `share:done` | `null` | Peer received the data (host side) |
| `share:received` | `{ title: string, state: State }` | Tab received and saved (guest side) |
| `share:error` | `string` — error message | Any failure on either side |

#### `tui/main.go` — share state machine

The TUI model gains three extra fields and three message types:

```go
// State
shareMode   shareMode        // shareOff | shareSending | shareReceive
shareCode   string           // generated code shown in legend bar
shareInput  string           // code being typed in receive mode
shareErr    string           // last error, shown briefly in legend
shareCancel context.CancelFunc

// Messages (returned by goroutine Cmds)
shareCodeMsg     { code string }
shareDoneMsg     {}
shareErrMsg      { err string }
shareReceivedMsg { title string; st core.State }
```

The `renderLegend()` function checks `shareMode` first and renders a contextual overlay instead of the normal shortcuts bar when a share operation is active.

---

## Privacy FAQ

**Does octoNote store my notes on any server?**
No. The relay server (`relay.magic-wormhole.io`) only handles the cryptographic handshake metadata (encrypted connection details). Your note content is encrypted client-side and transmitted directly between the two machines. The relay never stores anything.

**Who runs the relay server?**
The public relay is operated by the [Magic Wormhole open-source project](https://github.com/magic-wormhole/magic-wormhole). You can self-host a relay with `pip install magic-wormhole && twist wormhole-mailbox` and configure octoNote's `wormhole.Client.RendezvousURL` to point to it.

**Can a third party intercept the transfer if they have the code?**
Only if they enter the code before your intended peer does — and the SPAKE2 protocol will detect and reject a man-in-the-middle who provides the wrong code. The window of vulnerability is the time between you generating the code and your peer entering it (typically a few seconds). Don't generate a code and then leave it sitting for hours.

**What happens if the connection drops mid-transfer?**
The transfer fails with an error on both sides. Generate a new code and try again — no partial data is written.

**Is the share feature available without an internet connection?**
Not in the current implementation. The SPAKE2 handshake requires reaching the public relay server over the internet. LAN-only support (mDNS discovery) is a possible future addition.
