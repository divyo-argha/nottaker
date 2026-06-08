package main

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nottaker/nottaker/core"
)

// ─── Colour Palette ──────────────────────────────────────────────────────────

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
)

// ─── Styles ──────────────────────────────────────────────────────────────────

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
)

// ─── Bubble Tea Messages ──────────────────────────────────────────────────────

// savedMsg is sent by the storage layer (via a goroutine) to signal a flush.
type savedMsg struct{ at time.Time }

// ─── Model ───────────────────────────────────────────────────────────────────

type model struct {
	storage     *core.Storage
	state       core.State
	textareas   []textarea.Model
	width       int
	height      int
	lastSaved   time.Time
	dirty       bool        // true when unsaved changes exist
	quitting    bool
}

// ─── Init / Load ─────────────────────────────────────────────────────────────

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

	// Focus the active textarea.
	if m.state.ActiveIndex < len(m.textareas) {
		m.textareas[m.state.ActiveIndex].Focus()
	}
	return m
}

func newTextArea() textarea.Model {
	ta := textarea.New()
	ta.Placeholder = "Start typing…"
	ta.ShowLineNumbers = false
	ta.CharLimit = 0 // unlimited
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

// ─── Bubble Tea Interface ─────────────────────────────────────────────────────

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	// ── Terminal resize ──
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m = m.resizeTextAreas()

	// ── Keyboard ──
	case tea.KeyMsg:
		switch msg.Type {

		// Quit
		case tea.KeyCtrlC:
			m.syncSaveNow()
			m.quitting = true
			return m, tea.Quit

		// Next tab
		case tea.KeyCtrlRight, tea.KeyCtrlF:
			m = m.switchTab((m.state.ActiveIndex + 1) % len(m.state.Tabs))

		// Prev tab
		case tea.KeyCtrlLeft, tea.KeyCtrlB:
			idx := m.state.ActiveIndex - 1
			if idx < 0 {
				idx = len(m.state.Tabs) - 1
			}
			m = m.switchTab(idx)

		// Cycle with Tab key
		case tea.KeyTab:
			m = m.switchTab((m.state.ActiveIndex + 1) % len(m.state.Tabs))

		// New tab
		case tea.KeyCtrlN:
			m = m.newTab()
			m.triggerSave()

		// Close tab
		case tea.KeyCtrlW:
			m = m.closeTab()
			m.triggerSave()

		// All other keys → delegate to active textarea
		default:
			idx := m.state.ActiveIndex
			updated, cmd := m.textareas[idx].Update(msg)
			m.textareas[idx] = updated
			cmds = append(cmds, cmd)
			// Sync body to state and schedule async save.
			m.state.Tabs[idx].Body = m.textareas[idx].Value()
			m.state.Tabs[idx].CursorLine = m.textareas[idx].Line()
			m.state.Tabs[idx].UpdatedAt = time.Now()
			m.dirty = true
			m.triggerSave()
		}

	// ── Save acknowledgement ──
	case savedMsg:
		m.lastSaved = msg.at
		m.dirty = false
	}

	// Pass all other messages to the active textarea.
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

// ─── View ────────────────────────────────────────────────────────────────────

func (m model) View() string {
	if m.quitting {
		return styleTitle.Render("nottaker — bye! 👋") + "\n"
	}
	if m.width == 0 {
		return "Loading…"
	}

	var b strings.Builder

	// App title row
	title := styleTitle.Render("✦ nottaker")
	tabCount := styleTabCount.Render(fmt.Sprintf(" %d tab(s)", len(m.state.Tabs)))
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, title, tabCount))
	b.WriteString("\n")

	// Tab bar
	b.WriteString(m.renderTabBar())
	b.WriteString("\n")

	// Content area
	b.WriteString(m.renderContent())
	b.WriteString("\n")

	// Legend / status bar
	b.WriteString(m.renderLegend())

	return b.String()
}

func (m model) renderTabBar() string {
	tabs := make([]string, len(m.state.Tabs))
	for i, tab := range m.state.Tabs {
		label := truncate(tab.Title, 14)
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

	// Available height: total - title(1) - tabbar(3) - legend(1) - padding(2)
	contentH := m.height - 8
	if contentH < 4 {
		contentH = 4
	}
	contentW := m.width - 4 // account for border + padding

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
	shortcuts := []struct{ key, desc string }{
		{"^N", "new"},
		{"^W", "close"},
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
	if m.dirty {
		saveStatus = styleUnsaved.Render("● unsaved")
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

// ─── Tab Operations ───────────────────────────────────────────────────────────

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
	// Blur old active.
	m.textareas[m.state.ActiveIndex].Blur()
	m.state.ActiveIndex = len(m.state.Tabs) - 1
	return m
}

func (m model) closeTab() model {
	if len(m.state.Tabs) <= 1 {
		// Never close the last tab — just clear it.
		m.textareas[0].Reset()
		m.state.Tabs[0].Body = ""
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

// ─── Save Helpers ─────────────────────────────────────────────────────────────

// triggerSave queues an async save; the background goroutine handles the write.
func (m *model) triggerSave() {
	m.storage.Save(m.state)
}

// syncSaveNow is called on quit — drains via the channel (close + wait).
func (m *model) syncSaveNow() {
	m.storage.Save(m.state)
}

// ─── Layout ──────────────────────────────────────────────────────────────────

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

// ─── Utilities ────────────────────────────────────────────────────────────────

func truncate(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	return string(runes[:max-1]) + "…"
}

// visibleLen approximates the visible character width (strips ANSI codes naively).
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

// ─── Main ────────────────────────────────────────────────────────────────────

func main() {
	// Initialise storage.
	s, err := core.NewStorage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "nottaker: %v\n", err)
		os.Exit(1)
	}
	defer s.Close()

	// Load persisted state.
	st, err := s.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "nottaker: load state: %v\n", err)
		os.Exit(1)
	}

	// Build model and run.
	m := initialModel(s, st)
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),       // full-screen TUI
		tea.WithMouseCellMotion(), // optional: enable mouse clicks
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "nottaker: %v\n", err)
		os.Exit(1)
	}
}
