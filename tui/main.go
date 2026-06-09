package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nottaker/octonote/core"
)

const (
	colBg       = "#0f0f0f"
	colSurface  = "#1a1a2e"
	colBorder   = "#2d2d4e"
	colAccent   = "#7c3aed"
	colAccentLt = "#a78bfa"
	colMuted    = "#6b7280"
	colText     = "#e5e7eb"
	colSubtle   = "#9ca3af"
	colWarn     = "#f59e0b"
	colSuccess  = "#10b981"
	colTabBg    = "#16213e"
	colTabAct   = "#7c3aed"
	colErr      = "#ef4444"
)

var (
	styleTabInactive = lipgloss.NewStyle().
				Padding(0, 2).
				Background(lipgloss.Color(colTabBg)).
				Foreground(lipgloss.Color(colMuted)).
				Border(lipgloss.Border{
			Top: "─", Bottom: " ", Left: "│", Right: "│",
			TopLeft: "╭", TopRight: "╮", BottomLeft: " ", BottomRight: " ",
		}, true).
		BorderForeground(lipgloss.Color(colBorder))

	styleTabActive = lipgloss.NewStyle().
			Padding(0, 2).
			Background(lipgloss.Color(colAccent)).
			Foreground(lipgloss.Color("#ffffff")).
			Bold(true).
			Border(lipgloss.Border{
			Top: "─", Bottom: " ", Left: "│", Right: "│",
			TopLeft: "╭", TopRight: "╮", BottomLeft: " ", BottomRight: " ",
		}, true).
		BorderForeground(lipgloss.Color(colAccentLt))

	styleTabBar = lipgloss.NewStyle().
			Background(lipgloss.Color(colBg)).
			Padding(0, 1)

	styleContentBox = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(colAccent)).
				Padding(0, 1)

	styleContentBoxBlur = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(colBorder)).
				Padding(0, 1)

	styleLegend = lipgloss.NewStyle().
			Background(lipgloss.Color(colSurface)).
			Foreground(lipgloss.Color(colSubtle)).
			Padding(0, 1)

	styleKey = lipgloss.NewStyle().
			Background(lipgloss.Color(colAccent)).
			Foreground(lipgloss.Color("#ffffff")).
			Padding(0, 1).
			Bold(true)

	styleSaved = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colSuccess)).
			Bold(true)

	styleUnsaved = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colWarn))

	styleTitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colAccentLt)).
			Bold(true).
			Padding(0, 1)

	styleTabCount = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colSubtle))

	styleShareCode = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color(colAccent)).
			Bold(true).
			Padding(0, 2)

	styleShareInfo = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colAccentLt))

	styleShareErr = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colErr)).
			Bold(true)

	styleFilePrompt = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colText))

	styleFileInput = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#1e1e3f")).
			Padding(0, 1)

	styleFileErr = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colErr)).
			Bold(true)
)
			
type savedMsg struct{ at time.Time }

type shareDoneMsg struct{}
type shareCodeMsg struct{ code string }
type shareErrMsg struct{ err string }
type shareReceivedMsg struct {
	title string
	st    core.State
}
type shareStartedMsg struct {
	code string
	wait func() error
}
type shareWaitResultMsg struct {
	err error
}

type fileOpenedMsg struct {
	path    string
	content string
}
type fileSavedMsg struct {
	path string
	at   time.Time
}
type fileErrMsg struct{ err string }


type shareMode int

const (
	shareOff       shareMode = iota
	shareSending             // waiting for peer to connect
	shareReceive             // user typing the code
	shareReceiving           // receiver is connecting/handshaking
)

type filePromptMode int

const (
	filePromptOff     filePromptMode = iota
	filePromptOpen                   // user typing a path to open
	filePromptSave                   // user typing a path to save-as
	filePromptConfirm                // confirming close of dirty file (Y/N/Esc)
)


type model struct {
	storage   *core.Storage
	state     core.State
	textareas []textarea.Model
	width     int
	height    int
	lastSaved time.Time
	dirty     bool
	quitting  bool

	// share state
	shareMode   shareMode
	shareCode   string // generated code shown to sender
	shareInput  string // code being typed by receiver
	shareErr    string // last share error message
	shareCancel context.CancelFunc

	// file I/O state
	fileMode         filePromptMode
	fileInput        string // text being typed in the path/confirm prompt
	fileErr          string // last file I/O error message
	filePendingClose bool   // true when close-tab is waiting for a save confirmation
}

func initialModel(s *core.Storage, st core.State) model {
	tas := make([]textarea.Model, len(st.Tabs))
	for i, tab := range st.Tabs {
		tas[i] = newTextArea()
		tas[i].SetValue(tab.Body)
	}

	m := model{
		storage:   s,
		state:     st,
		textareas: tas,
		lastSaved: time.Now(),
	}

	if m.state.ActiveIndex < len(m.textareas) {
		m.textareas[m.state.ActiveIndex].Focus()
	}
	return m
}

func newTextArea() textarea.Model {
	ta := textarea.New()
	ta.Placeholder = "Start typing…"
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	ta.SetWidth(80)
	ta.SetHeight(20)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle().Background(lipgloss.Color("#1e1e3f"))
	ta.FocusedStyle.Base = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colText))
	ta.BlurredStyle.Base = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colMuted))
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color(colMuted))
	ta.BlurredStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color(colBorder))
	return ta
}

// ── Init ──────────────────────────────────────────────────────────────────────

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case shareDoneMsg:
		m.shareMode = shareOff
		m.shareCode = ""
		m.shareErr = ""

	case shareCodeMsg:
		m.shareCode = msg.code
		m.shareErr = ""

	case shareStartedMsg:
		m.shareCode = msg.code
		m.shareErr = ""
		return m, func() tea.Msg {
			err := msg.wait()
			return shareWaitResultMsg{err: err}
		}

	case shareWaitResultMsg:
		if m.shareMode != shareSending {
			// User cancelled in the meantime, ignore.
			return m, nil
		}
		m.shareMode = shareOff
		m.shareCode = ""
		if msg.err != nil {
			m.shareErr = msg.err.Error()
		} else {
			m.shareErr = ""
		}

	case shareErrMsg:
		m.shareMode = shareOff
		m.shareCode = ""
		m.shareErr = msg.err

	case shareReceivedMsg:
		m.state = msg.st
		m.shareMode = shareOff
		m.shareInput = ""
		m.shareErr = ""

		tas := make([]textarea.Model, len(m.state.Tabs))
		for i, tab := range m.state.Tabs {
			tas[i] = newTextArea()
			tas[i].SetValue(tab.Body)
		}
		m.textareas = tas
		if m.state.ActiveIndex < len(m.textareas) {
			m.textareas[m.state.ActiveIndex].Focus()
		}
		m = m.resizeTextAreas()

	// ── File I/O messages ─────────────────────────────────────────────────────
	case fileOpenedMsg:
		m.fileMode = filePromptOff
		m.fileInput = ""
		m.fileErr = ""
		m = m.loadFileIntoTab(msg.path, msg.content)
		m.triggerSave()

	case fileSavedMsg:
		m.fileMode = filePromptOff
		m.fileInput = ""
		m.fileErr = ""
		idx := m.state.ActiveIndex
		m.state.Tabs[idx].FilePath = msg.path
		m.state.Tabs[idx].FileIsDirty = false
		m.lastSaved = msg.at
		m.dirty = false
		m.triggerSave()
	
		if m.filePendingClose {
			m.filePendingClose = false
			m = m.closeTab()
			m.triggerSave()
		}

	case fileErrMsg:
		m.fileErr = msg.err
		m.fileMode = filePromptOff
		m.fileInput = ""
		m.filePendingClose = false

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m = m.resizeTextAreas()
	case savedMsg:
		m.lastSaved = msg.at
		m.dirty = false

	case tea.KeyMsg:

		if m.fileErr != "" {
			m.fileErr = ""
		}

		if m.fileMode == filePromptConfirm {
			switch strings.ToLower(msg.String()) {
			case "y":
				idx := m.state.ActiveIndex
				path := m.state.Tabs[idx].FilePath
				content := m.textareas[idx].Value()
				m.filePendingClose = true
				m.fileMode = filePromptOff
				cmds = append(cmds, func() tea.Msg {
					if err := core.SaveFile(path, content); err != nil {
						return fileErrMsg{err: err.Error()}
					}
					return fileSavedMsg{path: path, at: time.Now()}
				})
			case "n":
				m.fileMode = filePromptOff
				m.filePendingClose = false
				m = m.closeTab()
				m.triggerSave()
			default:
				if msg.Type == tea.KeyEscape || msg.Type == tea.KeyCtrlC {
					m.fileMode = filePromptOff
					m.filePendingClose = false
				}
			}
			return m, tea.Batch(cmds...)
		}

		if m.fileMode == filePromptOpen || m.fileMode == filePromptSave {
			switch msg.Type {
			case tea.KeyEscape, tea.KeyCtrlC:
				m.fileMode = filePromptOff
				m.fileInput = ""
				m.filePendingClose = false
			case tea.KeyEnter:
				path := strings.TrimSpace(m.fileInput)
				if path == "" {
					break
				}
				if m.fileMode == filePromptOpen {
					cmds = append(cmds, func() tea.Msg {
						content, err := core.OpenFile(path)
						if err != nil {
							return fileErrMsg{err: err.Error()}
						}
						return fileOpenedMsg{path: path, content: content}
					})
				} else {
					// save-as
					content := m.textareas[m.state.ActiveIndex].Value()
					pending := m.filePendingClose
					cmds = append(cmds, func() tea.Msg {
						_ = pending // captured for post-save close logic via filePendingClose field
						if err := core.SaveFile(path, content); err != nil {
							return fileErrMsg{err: err.Error()}
						}
						return fileSavedMsg{path: path, at: time.Now()}
					})
				}
			case tea.KeyBackspace:
				if len(m.fileInput) > 0 {
					runes := []rune(m.fileInput)
					m.fileInput = string(runes[:len(runes)-1])
				}
			default:
				if msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace {
					m.fileInput += msg.String()
				}
			}
			return m, tea.Batch(cmds...)
		}

		if m.shareMode == shareReceive || m.shareMode == shareReceiving {
			if m.shareMode == shareReceiving {
				switch msg.Type {
				case tea.KeyEscape, tea.KeyCtrlC:
					if m.shareCancel != nil {
						m.shareCancel()
					}
					m.shareMode = shareOff
					m.shareInput = ""
					m.shareErr = ""
				}
				return m, nil
			}

			switch msg.Type {
			case tea.KeyEscape, tea.KeyCtrlC:
				if m.shareCancel != nil {
					m.shareCancel()
				}
				m.shareMode = shareOff
				m.shareInput = ""
				m.shareErr = ""
			case tea.KeyEnter:
				code := strings.TrimSpace(m.shareInput)
				if code != "" {
					ctx, cancel := context.WithCancel(context.Background())
					m.shareCancel = cancel
					m.shareMode = shareReceiving
					s := m.storage
					st := m.state
					cmds = append(cmds, func() tea.Msg {
						res, err := core.ShareReceive(ctx, code)
						if err != nil {
							if ctx.Err() != nil {
								return nil
							}
							return shareErrMsg{err: err.Error()}
						}
						newTab := core.Tab{
							ID:        fmt.Sprintf("%x", time.Now().UnixNano()),
							Title:     res.TabTitle,
							Body:      res.Body,
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						}
						st.Tabs = append(st.Tabs, newTab)
						st.ActiveIndex = len(st.Tabs) - 1
						s.Save(st)
						return shareReceivedMsg{title: res.TabTitle, st: st}
					})
				}
			case tea.KeyBackspace:
				if len(m.shareInput) > 0 {
					m.shareInput = m.shareInput[:len(m.shareInput)-1]
				}
			case tea.KeyRunes:
				m.shareInput += msg.String()
			}
			return m, tea.Batch(cmds...)
		}

		// ── Normal mode ───────────────────────────────────────────────────────
		if m.shareErr != "" {
			m.shareErr = ""
		}

		switch msg.Type {

		case tea.KeyCtrlC:
			if m.shareMode == shareSending && m.shareCancel != nil {
				m.shareCancel()
				m.shareMode = shareOff
				m.shareCode = ""
				return m, nil
			}
			m.syncSaveNow()
			m.quitting = true
			return m, tea.Quit

		case tea.KeyCtrlRight, tea.KeyCtrlF:
			m = m.switchTab((m.state.ActiveIndex + 1) % len(m.state.Tabs))

		case tea.KeyCtrlLeft, tea.KeyCtrlB:
			idx := m.state.ActiveIndex - 1
			if idx < 0 {
				idx = len(m.state.Tabs) - 1
			}
			m = m.switchTab(idx)

		case tea.KeyTab:
			m = m.switchTab((m.state.ActiveIndex + 1) % len(m.state.Tabs))

		case tea.KeyCtrlN:
			m = m.newTab()
			m.triggerSave()

		case tea.KeyCtrlW:
			m, cmds = m.handleClose(cmds)

		case tea.KeyCtrlO:
			m.fileMode = filePromptOpen
			m.fileInput = ""

		default:
			if msg.Type == tea.KeyRunes && !msg.Alt {
				idx := m.state.ActiveIndex
				updated, cmd := m.textareas[idx].Update(msg)
				m.textareas[idx] = updated
				cmds = append(cmds, cmd)
				m.syncTabBody(idx)
				break
			}

			if msg.Type == tea.KeyCtrlS && !msg.Alt {
				m, cmds = m.handleSave(cmds)
				break
			}

			if msg.Type == tea.KeyCtrlR {
				m.shareMode = shareReceive
				m.shareInput = ""
				m.shareErr = ""
				break
			}

			idx := m.state.ActiveIndex
			updated, cmd := m.textareas[idx].Update(msg)
			m.textareas[idx] = updated
			cmds = append(cmds, cmd)
			m.syncTabBody(idx)
		}
	}

	if _, ok := msg.(tea.KeyMsg); !ok {
		idx := m.state.ActiveIndex
		if idx < len(m.textareas) {
			updated, cmd := m.textareas[idx].Update(msg)
			m.textareas[idx] = updated
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) handleClose(cmds []tea.Cmd) (model, []tea.Cmd) {
	idx := m.state.ActiveIndex
	tab := m.state.Tabs[idx]
	content := m.textareas[idx].Value()

	if tab.FilePath != "" && tab.FileIsDirty {
		m.fileMode = filePromptConfirm
		m.filePendingClose = true
		return m, cmds
	}

	if tab.FilePath == "" && strings.TrimSpace(content) != "" {
		m.fileMode = filePromptSave
		m.fileInput = ""
		m.filePendingClose = true
		return m, cmds
	}

	m = m.closeTab()
	m.triggerSave()
	return m, cmds
}

func (m model) handleSave(cmds []tea.Cmd) (model, []tea.Cmd) {
	idx := m.state.ActiveIndex
	tab := m.state.Tabs[idx]
	content := m.textareas[idx].Value()

	if tab.FilePath != "" {
		path := tab.FilePath
		cmds = append(cmds, func() tea.Msg {
			if err := core.SaveFile(path, content); err != nil {
				return fileErrMsg{err: err.Error()}
			}
			return fileSavedMsg{path: path, at: time.Now()}
		})
	} else {
		m.fileMode = filePromptSave
		m.fileInput = ""
	}
	return m, cmds
}

func (m *model) shareActiveTab(cmds []tea.Cmd) []tea.Cmd {
	if m.shareMode == shareSending {
		return cmds
	}
	ctx, cancel := context.WithCancel(context.Background())
	m.shareCancel = cancel
	m.shareMode = shareSending
	m.shareCode = "connecting…"
	m.shareErr = ""
	tab := m.state.Tabs[m.state.ActiveIndex]
	cmds = append(cmds, func() tea.Msg {
		code, wait, err := core.ShareSend(ctx, tab, "")
		if err != nil {
			return shareErrMsg{err: err.Error()}
		}
		return shareStartedMsg{code: code, wait: wait}
	})
	return cmds
}

func (m *model) syncTabBody(idx int) {
	m.state.Tabs[idx].Body = m.textareas[idx].Value()
	m.state.Tabs[idx].CursorLine = m.textareas[idx].Line()
	m.state.Tabs[idx].UpdatedAt = time.Now()
	if m.state.Tabs[idx].FilePath != "" {
		m.state.Tabs[idx].FileIsDirty = true
	}
	m.dirty = true
	m.triggerSave()
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m model) View() string {
	if m.quitting {
		return styleTitle.Render("✦ octonote — bye! 👋") + "\n"
	}
	if m.width == 0 {
		return "Loading…"
	}

	var b strings.Builder

	title := styleTitle.Render("✦ octonote")
	tabCount := styleTabCount.Render(fmt.Sprintf(" %d tab(s)", len(m.state.Tabs)))
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, title, tabCount))
	b.WriteString("\n")

	b.WriteString(m.renderTabBar())
	b.WriteString("\n")

	b.WriteString(m.renderContent())
	b.WriteString("\n")

	b.WriteString(m.renderLegend())

	return b.String()
}

func (m model) renderTabBar() string {
	tabs := make([]string, len(m.state.Tabs))
	for i, tab := range m.state.Tabs {
		label := truncate(tab.Title, 14)

		unsaved := (tab.FilePath == "" && strings.TrimSpace(m.textareas[i].Value()) != "") ||
			(tab.FilePath != "" && tab.FileIsDirty)
		if unsaved {
			label = "● " + label
		}

		if i == m.state.ActiveIndex {
			tabs[i] = styleTabActive.Render(fmt.Sprintf(" %d: %s ", i+1, label))
		} else {
			tabs[i] = styleTabInactive.Render(fmt.Sprintf(" %d: %s ", i+1, label))
		}
	}
	row := lipgloss.JoinHorizontal(lipgloss.Bottom, tabs...)
	return styleTabBar.Width(m.width).Render(row)
}

func (m model) renderContent() string {
	idx := m.state.ActiveIndex
	if idx >= len(m.textareas) {
		return ""
	}

	contentH := m.height - 8
	if contentH < 4 {
		contentH = 4
	}
	contentW := m.width - 4

	m.textareas[idx].SetWidth(contentW)
	m.textareas[idx].SetHeight(contentH)

	var box lipgloss.Style
	if m.textareas[idx].Focused() {
		box = styleContentBox
	} else {
		box = styleContentBoxBlur
	}

	return box.
		Width(m.width - 2).
		Render(m.textareas[idx].View())
}

func (m model) renderLegend() string {
	if m.fileErr != "" {
		return styleLegend.Width(m.width).Render(
			styleFileErr.Render("✗ " + m.fileErr),
		)
	}

	if m.fileMode == filePromptConfirm {
		msg := styleFileErr.Render("Unsaved changes!") +
			styleFilePrompt.Render("  Save before closing?  ") +
			styleKey.Render("Y") + " save  " +
			styleKey.Render("N") + " discard  " +
			styleKey.Render("Esc") + " cancel"
		return styleLegend.Width(m.width).Render(msg)
	}

	if m.fileMode == filePromptOpen {
		input := styleFileInput.Render(m.fileInput + "▌")
		prompt := styleFilePrompt.Render("Open: ") + input +
			styleFilePrompt.Render("  ") + styleKey.Render("↵") +
			styleFilePrompt.Render(" open  ") + styleKey.Render("Esc") + " cancel"
		return styleLegend.Width(m.width).Render(prompt)
	}

	if m.fileMode == filePromptSave {
		input := styleFileInput.Render(m.fileInput + "▌")
		prompt := styleFilePrompt.Render("Save as: ") + input +
			styleFilePrompt.Render("  ") + styleKey.Render("↵") +
			styleFilePrompt.Render(" save  ") + styleKey.Render("Esc") + " cancel"
		return styleLegend.Width(m.width).Render(prompt)
	}

	if m.shareMode == shareSending {
		var status string
		if m.shareCode == "connecting…" {
			status = styleShareInfo.Render("opening wormhole…")
		} else {
			status = "share code: " + styleShareCode.Render(m.shareCode) +
				styleShareInfo.Render("  waiting for peer…  ") +
				styleKey.Render("^C") + " cancel"
		}
		return styleLegend.Width(m.width).Render(status)
	}
	if m.shareMode == shareReceive {
		input := styleShareCode.Render("_" + m.shareInput + "_")
		prompt := styleShareInfo.Render("enter code: ") + input +
			styleShareInfo.Render("  then ") + styleKey.Render("↵") +
			styleShareInfo.Render(" to connect  ") + styleKey.Render("Esc") + " cancel"
		return styleLegend.Width(m.width).Render(prompt)
	}
	if m.shareMode == shareReceiving {
		status := styleShareInfo.Render("connecting to peer…  ") +
			styleKey.Render("Esc") + " cancel"
		return styleLegend.Width(m.width).Render(status)
	}
	if m.shareErr != "" {
		errMsg := styleShareErr.Render("share error: " + m.shareErr)
		return styleLegend.Width(m.width).Render(errMsg)
	}

	// ── Normal legend ─────────────────────────────────────────────────────────
	shortcuts := []struct{ key, desc string }{
		{"^N", "new"},
		{"^W", "close"},
		{"^O", "open"},
		{"^S", "save"},
		{"⇧^S", "share"},
		{"^R", "receive"},
		{"^→/←", "switch"},
		{"Tab", "cycle"},
		{"^C", "quit"},
	}

	var parts []string
	for _, s := range shortcuts {
		k := styleKey.Render(s.key)
		parts = append(parts, k+" "+s.desc)
	}

	var saveStatus string
	idx := m.state.ActiveIndex
	tab := m.state.Tabs[idx]
	if tab.FilePath != "" && !tab.FileIsDirty {
		saveStatus = styleSaved.Render("✓ " + filepath.Base(tab.FilePath))
	} else if tab.FilePath != "" && tab.FileIsDirty {
		saveStatus = styleUnsaved.Render("● " + filepath.Base(tab.FilePath) + " (unsaved)")
	} else if m.dirty {
		saveStatus = styleUnsaved.Render("● unsaved (^S to save)")
	} else {
		saveStatus = styleSaved.Render("✓ saved " + m.lastSaved.Format("15:04:05"))
	}

	left := strings.Join(parts, "  ")
	right := saveStatus

	gap := m.width - visibleLen(left) - visibleLen(right) - 4
	if gap < 1 {
		gap = 1
	}

	return styleLegend.
		Width(m.width).
		Render(left + strings.Repeat(" ", gap) + right)
}

func (m model) switchTab(idx int) model {
	if idx < 0 || idx >= len(m.state.Tabs) {
		return m
	}
	m.textareas[m.state.ActiveIndex].Blur()
	m.state.ActiveIndex = idx
	m.textareas[idx].Focus()
	m.triggerSave()
	return m
}

func (m model) newTab() model {
	title := fmt.Sprintf("tab %d", len(m.state.Tabs)+1)
	tab := core.NewTab(title)
	m.state.Tabs = append(m.state.Tabs, tab)
	ta := newTextArea()
	ta.Focus()
	m.textareas = append(m.textareas, ta)
	m.textareas[m.state.ActiveIndex].Blur()
	m.state.ActiveIndex = len(m.state.Tabs) - 1
	return m
}

func (m model) closeTab() model {
	if len(m.state.Tabs) <= 1 {
		m.textareas[0].Reset()
		m.state.Tabs[0].Body = ""
		m.state.Tabs[0].FilePath = ""
		m.state.Tabs[0].FileIsDirty = false
		m.state.Tabs[0].UpdatedAt = time.Now()
		return m
	}
	idx := m.state.ActiveIndex
	m.state.Tabs = append(m.state.Tabs[:idx], m.state.Tabs[idx+1:]...)
	m.textareas = append(m.textareas[:idx], m.textareas[idx+1:]...)
	if m.state.ActiveIndex >= len(m.state.Tabs) {
		m.state.ActiveIndex = len(m.state.Tabs) - 1
	}
	m.textareas[m.state.ActiveIndex].Focus()
	return m
}

func (m model) loadFileIntoTab(path, content string) model {
	idx := m.state.ActiveIndex
	currentContent := strings.TrimSpace(m.textareas[idx].Value())

	if currentContent == "" && m.state.Tabs[idx].FilePath == "" {
		m.state.Tabs[idx].Title = filepath.Base(path)
		m.state.Tabs[idx].Body = content
		m.state.Tabs[idx].FilePath = path
		m.state.Tabs[idx].FileIsDirty = false
		m.state.Tabs[idx].UpdatedAt = time.Now()
		m.textareas[idx].SetValue(content)
	} else {
		tab := core.NewTab(filepath.Base(path))
		tab.Body = content
		tab.FilePath = path
		tab.FileIsDirty = false
		m.state.Tabs = append(m.state.Tabs, tab)
		ta := newTextArea()
		ta.SetValue(content)
		ta.Focus()
		m.textareas = append(m.textareas, ta)
		m.textareas[m.state.ActiveIndex].Blur()
		m.state.ActiveIndex = len(m.state.Tabs) - 1
	}
	return m
}

func (m *model) triggerSave() {
	m.storage.Save(m.state)
}

func (m *model) syncSaveNow() {
	m.storage.Save(m.state)
}

func (m model) resizeTextAreas() model {
	contentH := m.height - 8
	if contentH < 4 {
		contentH = 4
	}
	contentW := m.width - 4
	for i := range m.textareas {
		m.textareas[i].SetWidth(contentW)
		m.textareas[i].SetHeight(contentH)
	}
	return m
}

func truncate(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	return string(runes[:max-1]) + "…"
}

func visibleLen(s string) int {
	inEscape := false
	count := 0
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		count++
	}
	return count
}

func main() {
	s, err := core.NewStorage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "octonote: %v\n", err)
		os.Exit(1)
	}
	defer s.Close()

	st, err := s.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "octonote: load state: %v\n", err)
		os.Exit(1)
	}

	m := initialModel(s, st)
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "octonote: %v\n", err)
		os.Exit(1)
	}
}
