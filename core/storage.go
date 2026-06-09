package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Tab struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	CursorLine  int       `json:"cursor_line"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	FilePath    string    `json:"file_path,omitempty"`    // absolute path to on-disk file, empty if unsaved
	FileIsDirty bool      `json:"file_is_dirty,omitempty"` // true when body differs from last file save
}

type State struct {
	Tabs        []Tab `json:"tabs"`
	ActiveIndex int   `json:"active_index"`
	Version     int   `json:"version"`
}

type Storage struct {
	mu      sync.Mutex
	dir     string
	file    string
	ch      chan State
	lastErr error
	done    chan struct{}
}

func StateDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("octonote: cannot determine config dir: %w", err)
	}
	return filepath.Join(base, "octonote"), nil
}

func NewStorage() (*Storage, error) {
	dir, err := StateDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("octonote: cannot create config dir %s: %w", dir, err)
	}

	s := &Storage{
		dir:  dir,
		file: filepath.Join(dir, "state.json"),
		ch:   make(chan State, 64),
		done: make(chan struct{}),
	}

	go s.syncWriter()
	return s, nil
}

func (s *Storage) Load() (State, error) {
	data, err := os.ReadFile(s.file)
	if os.IsNotExist(err) {
		return defaultState(), nil
	}
	if err != nil {
		return State{}, fmt.Errorf("octonote: read state: %w", err)
	}

	var st State
	if err := json.Unmarshal(data, &st); err != nil {
		return defaultState(), nil
	}

	if len(st.Tabs) == 0 {
		st = defaultState()
	}
	if st.ActiveIndex >= len(st.Tabs) || st.ActiveIndex < 0 {
		st.ActiveIndex = 0
	}
	return st, nil
}

func (s *Storage) Save(st State) {
	select {
	case s.ch <- st:
	default:
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

func (s *Storage) Close() {
	close(s.ch)
	<-s.done
}

func (s *Storage) LastError() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastErr
}

func (s *Storage) Dir() string { return s.dir }

func (s *Storage) syncWriter() {
	defer close(s.done)

	for st := range s.ch {
		drained := st
		for {
			select {
			case newer, ok := <-s.ch:
				if !ok {
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

func (s *Storage) atomicWrite(st State) {
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		s.setErr(err)
		return
	}

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

func generateID() string {
	return fmt.Sprintf("%x", time.Now().UnixNano())
}
