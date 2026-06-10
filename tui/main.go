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
	"github.com/ncruces/zenity"
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

// ── Messages ──────────────────────────────────────────────────────────────────

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
type shareWaitResultMsg struct{ err error }

type fileOpenedMsg struct {
	path    string
	content string
}
type fileSavedMsg struct {
	path string
	at   time.Time
}
type fileErrMsg struct{ err string }

type zenityFileSelectedMsg struct{ path string }
type zenityFileSaveSelectedMsg struct{ path string }
type zenityCanceledMsg struct{ mode filePromptMode }
type zenityFailedMsg struct {
	err  error
	mode filePromptMode
}

func selectFileCmd() tea.Msg {
	path, err := zenity.SelectFile(
		zenity.Title("Open File"),
		zenity.FileFilters{
			{
				Name: "Text Files",
				Patterns: []string{
					"*.txt", "*.md", "*.html", "*.json", "*.xml",
					"*.js", "*.ts", "*.css", "*.scss", "*.less",
					"*.go", "*.py", "*.sh", "*.bat", "*.ps1",
					"*.yaml", "*.yml", "*.ini", "*.conf", "*.cfg",
					"*.csv", "*.tsv", "*.log", "*.sql",
				},
				CaseFold: true,
			},
			{
				Name:     "All Files",
				Patterns: []string{"*"},
			},
		},
	)
	if err != nil {
		if err == zenity.ErrCanceled {
			return zenityCanceledMsg{mode: filePromptOpen}
		}
		return zenityFailedMsg{err: err, mode: filePromptOpen}
	}
	return zenityFileSelectedMsg{path: path}
}

func selectFileSaveCmd() tea.Msg {
	path, err := zenity.SelectFileSave(
		zenity.Title("Save File"),
		zenity.FileFilters{
			{
				Name: "Text Files",
				Patterns: []string{
					"*.txt", "*.md", "*.html", "*.json", "*.xml",
					"*.js", "*.ts", "*.css", "*.scss", "*.less",
					"*.go", "*.py", "*.sh", "*.bat", "*.ps1",
					"*.yaml", "*.yml", "*.ini", "*.conf", "*.cfg",
					"*.csv", "*.tsv", "*.log", "*.sql",
				},
				CaseFold: true,
			},
			{
				Name:     "All Files",
				Patterns: []string{"*"},
			},
		},
	)
	if err != nil {
		if err == zenity.ErrCanceled {
			return zenityCanceledMsg{mode: filePromptSave}
		}
		return zenityFailedMsg{err: err, mode: filePromptSave}
	}
	return zenityFileSaveSelectedMsg{path: path}
}

// ── Mode enums ────────────────────────────────────────────────────────────────

type shareMode int

const (
	shareOff       shareMode = iota
	shareSending             // waiting for peer to connect
	shareReceive             // user typing the wormhole code
	shareReceiving           // receiver connecting/handshaking
)

type filePromptMode int

const (
	filePromptOff     filePromptMode = iota
	filePromptOpen                   // user typing a path to open
	filePromptSave                   // user typing a path to save-as
	filePromptConfirm                // Y/N/Esc: save before close?
)

// ── Model ─────────────────────────────────────────────────────────────────────

type model struct {
	storage  *core.Storage
	state    core.State
	textareas []textarea.Model
	width    int
	height   int
	lastSaved time.Time
	dirty    bool
	quitting bool

	// share state
	shareMode   shareMode
	shareCode   string
	shareInput  string
	shareErr    string
	shareCancel context.CancelFunc

	// file I/O state
	fileMode         filePromptMode
	fileInput        string // typed path or Y/N
	fileErr          string
	filePendingClose bool // waiting for save before closing tab
	fileSubmitting   bool // async op in flight; block further edits
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
	ta.FocusedStyle.Base = lipgloss.NewStyle().Foreground(lipgloss.Color(colText))
	ta.BlurredStyle.Base = lipgloss.NewStyle().Foreground(lipgloss.Color(colMuted))
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color(colMuted))
	ta.BlurredStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color(colBorder))
	return ta
}

func (m model) Init() tea.Cmd { return textarea.Blink }

// ── Update ────────────────────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	// ── Async results ─────────────────────────────────────────────────────────

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
		return m, func() tea.Msg { return shareWaitResultMsg{err: msg.wait()} }

	case shareWaitResultMsg:
		if m.shareMode != shareSending {
			// User cancelled — ignore.
			return m, nil
		}
		m.shareMode = shareOff
		m.shareCode = ""
		if msg.err != nil {
			m.shareErr = msg.err.Error()
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

	case zenityFileSelectedMsg:
		m.fileInput = msg.path
		m.fileSubmitting = true
		cmds = append(cmds, func() tea.Msg {
			content, err := core.OpenFile(msg.path)
			if err != nil {
				return fileErrMsg{err: err.Error()}
			}
			return fileOpenedMsg{path: msg.path, content: content}
		})

	case zenityFileSaveSelectedMsg:
		m.fileInput = msg.path
		m.fileSubmitting = true
		content := m.textareas[m.state.ActiveIndex].Value()
		cmds = append(cmds, func() tea.Msg {
			if err := core.SaveFile(msg.path, content); err != nil {
				return fileErrMsg{err: err.Error()}
			}
			return fileSavedMsg{path: msg.path, at: time.Now()}
		})

	case zenityCanceledMsg:
		m.fileMode = filePromptOff
		m.fileInput = ""
		m.fileSubmitting = false
		m.filePendingClose = false

	case zenityFailedMsg:
		m.fileSubmitting = false
		m.fileErr = fmt.Sprintf("System dialog error: %v. Please enter path manually.", msg.err)

	case fileOpenedMsg:
		m.fileMode = filePromptOff
		m.fileInput = ""
		m.fileErr = ""
		m.fileSubmitting = false
		m = m.loadFileIntoTab(msg.path, msg.content)
		m.triggerSave()

	case fileSavedMsg:
		m.fileMode = filePromptOff
		m.fileInput = ""
		m.fileErr = ""
		m.fileSubmitting = false
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
		m.fileSubmitting = false
		m.filePendingClose = false

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m = m.resizeTextAreas()

	case savedMsg:
		m.lastSaved = msg.at
		m.dirty = false

	// ── Key handling ──────────────────────────────────────────────────────────
	case tea.KeyMsg:
		return m.handleKey(msg, cmds)
	}

	// Propagate non-key messages (blink tick, etc.) to active textarea.
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

// handleKey is the single entry point for all keyboard input.
// It is extracted to keep Update clean and to avoid break/fallthrough confusion.
func (m model) handleKey(msg tea.KeyMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	// Clear stale errors on any key.
	m.fileErr = ""
	m.shareErr = ""

	// ── File close-confirm prompt (Y / N / Esc) ───────────────────────────────
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

	// ── File path prompts (Open / Save-as) ───────────────────────────────────
	if m.fileMode == filePromptOpen || m.fileMode == filePromptSave {
		// If we're waiting for the async op (fileSubmitting), ignore all keys
		// except Escape so the user can't double-submit or corrupt state.
		if m.fileSubmitting {
			if msg.Type == tea.KeyEscape || msg.Type == tea.KeyCtrlC {
				m.fileMode = filePromptOff
				m.fileInput = ""
				m.fileSubmitting = false
				m.filePendingClose = false
			}
			return m, tea.Batch(cmds...)
		}

		switch msg.Type {
		case tea.KeyEscape, tea.KeyCtrlC:
			m.fileMode = filePromptOff
			m.fileInput = ""
			m.filePendingClose = false

		case tea.KeyEnter:
			path := strings.TrimSpace(m.fileInput)
			if path == "" {
				return m, tea.Batch(cmds...)
			}
			mode := m.fileMode
			content := m.textareas[m.state.ActiveIndex].Value()
			m.fileSubmitting = true
			cmds = append(cmds, func() tea.Msg {
				if mode == filePromptOpen {
					content, err := core.OpenFile(path)
					if err != nil {
						return fileErrMsg{err: err.Error()}
					}
					return fileOpenedMsg{path: path, content: content}
				}
				// Save-as
				if err := core.SaveFile(path, content); err != nil {
					return fileErrMsg{err: err.Error()}
				}
				return fileSavedMsg{path: path, at: time.Now()}
			})
			// fileMode stays set; fileSubmitting prevents further edits until
			// the async result resets everything.

		case tea.KeyBackspace, tea.KeyCtrlH:
			// Ctrl+H is the ASCII backspace (0x08) sent by many terminals.
			if len(m.fileInput) > 0 {
				runes := []rune(m.fileInput)
				m.fileInput = string(runes[:len(runes)-1])
			}

		case tea.KeyCtrlW:
			// Delete the last word (like readline's Ctrl+W).
			m.fileInput = deleteLastWord(m.fileInput)

		case tea.KeyCtrlU:
			// Clear the entire input line (like readline's Ctrl+U).
			m.fileInput = ""

		default:
			// Accept all printable runes including / . ~ - _
			if msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace {
				m.fileInput += msg.String()
			}
		}
		return m, tea.Batch(cmds...)
	}

	// ── Wormhole receive mode: user typing a share code ───────────────────────
	if m.shareMode == shareReceiving {
		if msg.Type == tea.KeyEscape || msg.Type == tea.KeyCtrlC {
			if m.shareCancel != nil {
				m.shareCancel()
			}
			m.shareMode = shareOff
			m.shareInput = ""
		}
		return m, tea.Batch(cmds...)
	}
	if m.shareMode == shareReceive {
		switch msg.Type {
		case tea.KeyEscape, tea.KeyCtrlC:
			if m.shareCancel != nil {
				m.shareCancel()
			}
			m.shareMode = shareOff
			m.shareInput = ""
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

	// ── Normal mode ───────────────────────────────────────────────────────────

	switch msg.Type {

	case tea.KeyCtrlC:
		// Cancel sharing if active; otherwise quit.
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

	case tea.KeyCtrlN, tea.KeyF5:
		m = m.newTab()
		m.triggerSave()

	// Ctrl+X or F4 → close tab (Ctrl+W is intercepted by macOS Terminal.app)
	case tea.KeyCtrlX, tea.KeyF4:
		m, cmds = m.handleClose(cmds)

	// Ctrl+O or F3 → open file
	case tea.KeyCtrlO, tea.KeyF3:
		m.fileMode = filePromptOpen
		m.fileInput = ""
		m.fileSubmitting = true
		cmds = append(cmds, selectFileCmd)

	// Ctrl+S or F2 → save to disk
	case tea.KeyCtrlS, tea.KeyF2:
		m, cmds = m.handleSave(cmds)

	// Ctrl+T → share active tab via wormhole (T for Transfer)
	case tea.KeyCtrlT:
		cmds = m.doShare(cmds)

	// Ctrl+R → receive from wormhole
	case tea.KeyCtrlR:
		m.shareMode = shareReceive
		m.shareInput = ""

	default:
		// All other keys (printable runes, arrows, etc.) go to the active textarea.
		idx := m.state.ActiveIndex
		updated, cmd := m.textareas[idx].Update(msg)
		m.textareas[idx] = updated
		cmds = append(cmds, cmd)
		if msg.Type == tea.KeyRunes || msg.Type == tea.KeyBackspace ||
			msg.Type == tea.KeyDelete || msg.Type == tea.KeyEnter {
			m.syncTabBody(idx)
		}
	}

	return m, tea.Batch(cmds...)
}

// ── File / share helpers ──────────────────────────────────────────────────────

// handleClose implements Ctrl+X (close tab).
// Prompts before closing if there are unsaved-to-disk changes.
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
		// New tab with content — ask where to save before closing using native dialog.
		m.fileMode = filePromptSave
		m.fileInput = ""
		m.filePendingClose = true
		m.fileSubmitting = true
		return m, append(cmds, selectFileSaveCmd)
	}

	m = m.closeTab()
	m.triggerSave()
	return m, cmds
}

// handleSave implements Ctrl+S.
func (m model) handleSave(cmds []tea.Cmd) (model, []tea.Cmd) {
	idx := m.state.ActiveIndex
	path := m.state.Tabs[idx].FilePath
	content := m.textareas[idx].Value()

	if path != "" {
		// Known file → overwrite silently.
		cmds = append(cmds, func() tea.Msg {
			if err := core.SaveFile(path, content); err != nil {
				return fileErrMsg{err: err.Error()}
			}
			return fileSavedMsg{path: path, at: time.Now()}
		})
	} else {
		// New tab → prompt for destination path using native system dialog.
		m.fileMode = filePromptSave
		m.fileInput = ""
		m.fileSubmitting = true
		cmds = append(cmds, selectFileSaveCmd)
	}
	return m, cmds
}

// doShare starts a Magic Wormhole send (Ctrl+T).
func (m model) doShare(cmds []tea.Cmd) []tea.Cmd {
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

// syncTabBody copies the textarea value back to the state and marks dirty.
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
		// Show ● when content is unsaved to disk.
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
	return box.Width(m.width - 2).Render(m.textareas[idx].View())
}

func (m model) renderLegend() string {
	// File I/O error banner.
	if m.fileErr != "" {
		return styleLegend.Width(m.width).Render(styleFileErr.Render("✗ " + m.fileErr))
	}

	// Close-confirm prompt.
	if m.fileMode == filePromptConfirm {
		msg := styleFileErr.Render("Unsaved changes!") +
			styleFilePrompt.Render("  Save before closing?  ") +
			styleKey.Render("Y") + " save  " +
			styleKey.Render("N") + " discard  " +
			styleKey.Render("Esc") + " cancel"
		return styleLegend.Width(m.width).Render(msg)
	}

	// Open-file prompt.
	if m.fileMode == filePromptOpen {
		if m.fileSubmitting {
			statusText := "Opening system file picker…"
			if m.fileInput != "" {
				statusText = "Opening " + m.fileInput + " …"
			}
			return styleLegend.Width(m.width).Render(
				styleFilePrompt.Render(statusText),
			)
		}
		input := styleFileInput.Render(m.fileInput + "▌")
		prompt := styleFilePrompt.Render("Open: ") + input +
			styleFilePrompt.Render("  ") + styleKey.Render("↵") +
			styleFilePrompt.Render(" open  ") + styleKey.Render("^U") + " clear  " +
			styleKey.Render("^W") + " del-word  " +
			styleKey.Render("Esc") + " cancel"
		return styleLegend.Width(m.width).Render(prompt)
	}

	// Save-as prompt.
	if m.fileMode == filePromptSave {
		if m.fileSubmitting {
			statusText := "Opening system save dialog…"
			if m.fileInput != "" {
				statusText = "Saving " + m.fileInput + " …"
			}
			return styleLegend.Width(m.width).Render(
				styleFilePrompt.Render(statusText),
			)
		}
		input := styleFileInput.Render(m.fileInput + "▌")
		prompt := styleFilePrompt.Render("Save as: ") + input +
			styleFilePrompt.Render("  ") + styleKey.Render("↵") +
			styleFilePrompt.Render(" save  ") + styleKey.Render("^U") + " clear  " +
			styleKey.Render("^W") + " del-word  " +
			styleKey.Render("Esc") + " cancel"
		return styleLegend.Width(m.width).Render(prompt)
	}

	// Share overlays.
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
		return styleLegend.Width(m.width).Render(
			styleShareInfo.Render("connecting to peer…  ") + styleKey.Render("Esc") + " cancel",
		)
	}
	if m.shareErr != "" {
		return styleLegend.Width(m.width).Render(styleShareErr.Render("share error: " + m.shareErr))
	}

	// Normal legend.
	shortcuts := []struct{ key, desc string }{
		{"^N/F5", "new"},
		{"^X/F4", "close"},
		{"^O/F3", "open"},
		{"^S/F2", "save"},
		{"^T", "share"},
		{"^R", "receive"},
		{"^→/←", "switch"},
		{"Tab", "cycle"},
		{"^C", "quit"},
	}
	var parts []string
	for _, s := range shortcuts {
		parts = append(parts, styleKey.Render(s.key)+" "+s.desc)
	}

	// Right-side save status.
	idx := m.state.ActiveIndex
	tab := m.state.Tabs[idx]
	var saveStatus string
	switch {
	case tab.FilePath != "" && !tab.FileIsDirty:
		saveStatus = styleSaved.Render("✓ " + filepath.Base(tab.FilePath))
	case tab.FilePath != "" && tab.FileIsDirty:
		saveStatus = styleUnsaved.Render("● " + filepath.Base(tab.FilePath) + " (unsaved)")
	case m.dirty:
		saveStatus = styleUnsaved.Render("● unsaved  (^S/F2 to save)")
	default:
		saveStatus = styleSaved.Render("✓ saved " + m.lastSaved.Format("15:04:05"))
	}

	left := strings.Join(parts, "  ")
	gap := m.width - visibleLen(left) - visibleLen(saveStatus) - 4
	if gap < 1 {
		gap = 1
	}
	return styleLegend.Width(m.width).Render(left + strings.Repeat(" ", gap) + saveStatus)
}

// ── Tab helpers ───────────────────────────────────────────────────────────────

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

// loadFileIntoTab puts file content into the current tab (if empty/new) or a new tab.
func (m model) loadFileIntoTab(path, content string) model {
	idx := m.state.ActiveIndex
	if strings.TrimSpace(m.textareas[idx].Value()) == "" && m.state.Tabs[idx].FilePath == "" {
		// Reuse current tab.
		m.state.Tabs[idx].Title = filepath.Base(path)
		m.state.Tabs[idx].Body = content
		m.state.Tabs[idx].FilePath = path
		m.state.Tabs[idx].FileIsDirty = false
		m.state.Tabs[idx].UpdatedAt = time.Now()
		m.textareas[idx].SetValue(content)
	} else {
		// Open in a new tab.
		tab := core.NewTab(filepath.Base(path))
		tab.Body = content
		tab.FilePath = path
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

func (m *model) triggerSave()  { m.storage.Save(m.state) }
func (m *model) syncSaveNow() { m.storage.Save(m.state) }

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

// ── Utilities ─────────────────────────────────────────────────────────────────

func truncate(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	return string([]rune(s)[:max-1]) + "…"
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

// deleteLastWord removes the last whitespace-delimited word from s,
// matching readline's Ctrl+W behaviour.
func deleteLastWord(s string) string {
	runes := []rune(strings.TrimRight(s, " \t"))
	// walk backwards over the last word
	i := len(runes) - 1
	for i >= 0 && runes[i] != ' ' && runes[i] != '/' && runes[i] != '\\' {
		i--
	}
	if i < 0 {
		return ""
	}
	return string(runes[:i+1])
}

// ── Main ──────────────────────────────────────────────────────────────────────

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
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "octonote: %v\n", err)
		os.Exit(1)
	}
}
