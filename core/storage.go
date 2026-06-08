package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ─── Data Model ─────────────────────────────────────────────────────────────

// Tab represents a single scratchpad tab with all its persisted state.
type Tab struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Body       string    `json:"body"`
	CursorLine int       `json:"cursor_line"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// State is the full persisted workspace state written to disk.
type State struct {
	Tabs        []Tab `json:"tabs"`
	ActiveIndex int   `json:"active_index"`
	Version     int   `json:"version"` // schema version for future migrations
}

// ─── Storage ─────────────────────────────────────────────────────────────────

// Storage manages reading and writing workspace state asynchronously.
// A single background goroutine owns all disk I/O so callers never block.
type Storage struct {
	mu      sync.Mutex
	dir     string        // config directory (contains state.json)
	file    string        // absolute path to state.json
	ch      chan State     // channel to the background writer goroutine
	lastErr error         // last background write error (informational only)
	done    chan struct{}  // closed when background goroutine exits
}

// StateDir returns the platform-specific config directory for nottaker.
// macOS  → ~/Library/Application Support/nottaker
// Linux  → ~/.config/nottaker   (respects XDG_CONFIG_HOME)
// Windows→ %AppData%/nottaker
func StateDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("nottaker: cannot determine config dir: %w", err)
	}
	return filepath.Join(base, "nottaker"), nil
}

// NewStorage creates a Storage instance, initialises the config directory,
// and starts the background writer goroutine.
func NewStorage() (*Storage, error) {
	dir, err := StateDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("nottaker: cannot create config dir %s: %w", dir, err)
	}

	s := &Storage{
		dir:  dir,
		file: filepath.Join(dir, "state.json"),
		ch:   make(chan State, 64), // generous buffer so keystrokes never block
		done: make(chan struct{}),
	}

	go s.syncWriter()
	return s, nil
}

// Load reads the persisted state from disk.
// If no state file exists (first run), a default single-tab state is returned.
func (s *Storage) Load() (State, error) {
	data, err := os.ReadFile(s.file)
	if os.IsNotExist(err) {
		return defaultState(), nil
	}
	if err != nil {
		return State{}, fmt.Errorf("nottaker: read state: %w", err)
	}

	var st State
	if err := json.Unmarshal(data, &st); err != nil {
		// Corrupted file — return default and let next save overwrite it.
		return defaultState(), nil
	}

	// Ensure at least one tab always exists.
	if len(st.Tabs) == 0 {
		st = defaultState()
	}
	// Clamp active index.
	if st.ActiveIndex >= len(st.Tabs) || st.ActiveIndex < 0 {
		st.ActiveIndex = 0
	}
	return st, nil
}

// Save queues a non-blocking asynchronous write of the given state.
// It returns immediately; the actual disk write happens in the background.
// Dropping a write on a full channel is acceptable — the most recent state
// will eventually be flushed on the next Save call or on Close.
func (s *Storage) Save(st State) {
	select {
	case s.ch <- st:
	default:
		// Channel full: discard the oldest pending write and try again.
		select {
		case <-s.ch:
		default:
		}
		select {
		case s.ch <- st:
		default:
		}
	}
}

// Close drains any remaining writes and stops the background goroutine.
// Always call this before the process exits.
func (s *Storage) Close() {
	close(s.ch)
	<-s.done
}

// LastError returns the most recent background write error (if any).
func (s *Storage) LastError() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastErr
}

// Dir returns the storage directory path (useful for debug display).
func (s *Storage) Dir() string { return s.dir }

// ─── Background goroutine ─────────────────────────────────────────────────

// syncWriter is the single goroutine responsible for all disk I/O.
// It reads from the channel and performs an atomic write (temp file + rename)
// so that a crash mid-write never corrupts the existing state file.
func (s *Storage) syncWriter() {
	defer close(s.done)

	for st := range s.ch {
		// Drain: take the latest item if more are queued.
		drained := st
		for {
			select {
			case newer, ok := <-s.ch:
				if !ok {
					// Channel closed while draining — flush the latest we have.
					s.atomicWrite(drained)
					return
				}
				drained = newer
			default:
				goto write
			}
		}
	write:
		s.atomicWrite(drained)
	}
}

// atomicWrite marshals state and writes it via a temp file + rename.
func (s *Storage) atomicWrite(st State) {
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		s.setErr(err)
		return
	}

	// Write to a temp file in the same directory (same filesystem → rename is atomic).
	tmp := s.file + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		s.setErr(err)
		return
	}
	if err := os.Rename(tmp, s.file); err != nil {
		s.setErr(err)
		return
	}
	s.setErr(nil)
}

func (s *Storage) setErr(err error) {
	s.mu.Lock()
	s.lastErr = err
	s.mu.Unlock()
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// defaultState returns a brand-new workspace with one empty tab.
func defaultState() State {
	now := time.Now()
	return State{
		Version:     1,
		ActiveIndex: 0,
		Tabs: []Tab{
			{
				ID:        generateID(),
				Title:     "scratch",
				Body:      "",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}
}

// NewTab creates and returns a new blank Tab.
func NewTab(title string) Tab {
	now := time.Now()
	return Tab{
		ID:        generateID(),
		Title:     title,
		Body:      "",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// generateID produces a unique tab identifier from the current nanosecond timestamp.
func generateID() string {
	return fmt.Sprintf("%x", time.Now().UnixNano())
}
