package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nottaker/nottaker/core"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the main application struct whose exported methods are bound to the
// Wails JS bridge and callable from the frontend as window.go.main.App.*
type App struct {
	ctx     context.Context
	storage *core.Storage
	state   core.State
	mu      sync.Mutex // guards state during concurrent JS calls
}

// NewApp creates a new App with the given storage backend.
func NewApp(storage *core.Storage) *App {
	return &App{storage: storage}
}

// ─── JS-Bound Methods ─────────────────────────────────────────────────────────

// GetState returns the full current workspace state as a JSON-serialisable value.
// Called by the frontend on startup.
func (a *App) GetState() core.State {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.state
}

// SaveTab updates the body and cursor position for a tab and triggers an async save.
// Called on every keypress from the frontend (debounced ~50 ms by JS).
func (a *App) SaveTab(index int, body string, cursorLine int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if index < 0 || index >= len(a.state.Tabs) {
		return
	}
	a.state.Tabs[index].Body = body
	a.state.Tabs[index].CursorLine = cursorLine
	a.state.Tabs[index].UpdatedAt = time.Now()
	a.storage.Save(a.state)
}

// NewTab creates a new blank tab, appends it, makes it active, and persists.
func (a *App) NewTab() core.State {
	a.mu.Lock()
	defer a.mu.Unlock()
	title := fmt.Sprintf("tab %d", len(a.state.Tabs)+1)
	tab := core.NewTab(title)
	a.state.Tabs = append(a.state.Tabs, tab)
	a.state.ActiveIndex = len(a.state.Tabs) - 1
	a.storage.Save(a.state)
	return a.state
}

// CloseTab removes the tab at index. If it is the last tab, the body is cleared
// instead. Returns the updated state.
func (a *App) CloseTab(index int) core.State {
	a.mu.Lock()
	defer a.mu.Unlock()
	if index < 0 || index >= len(a.state.Tabs) {
		return a.state
	}
	if len(a.state.Tabs) == 1 {
		// Cannot destroy last tab — clear it.
		a.state.Tabs[0].Body = ""
		a.state.Tabs[0].UpdatedAt = time.Now()
		a.storage.Save(a.state)
		return a.state
	}
	a.state.Tabs = append(a.state.Tabs[:index], a.state.Tabs[index+1:]...)
	if a.state.ActiveIndex >= len(a.state.Tabs) {
		a.state.ActiveIndex = len(a.state.Tabs) - 1
	}
	a.storage.Save(a.state)
	return a.state
}

// SetActiveTab switches the active tab index and persists.
func (a *App) SetActiveTab(index int) core.State {
	a.mu.Lock()
	defer a.mu.Unlock()
	if index < 0 || index >= len(a.state.Tabs) {
		return a.state
	}
	a.state.ActiveIndex = index
	a.storage.Save(a.state)
	return a.state
}

// RenameTab sets a new title for the tab at index and persists.
func (a *App) RenameTab(index int, title string) core.State {
	a.mu.Lock()
	defer a.mu.Unlock()
	if index < 0 || index >= len(a.state.Tabs) {
		return a.state
	}
	if title == "" {
		title = fmt.Sprintf("tab %d", index+1)
	}
	a.state.Tabs[index].Title = title
	a.state.Tabs[index].UpdatedAt = time.Now()
	a.storage.Save(a.state)
	return a.state
}

// GetStorageDir returns the config directory path (used for debugging in UI).
func (a *App) GetStorageDir() string {
	return a.storage.Dir()
}

// GetLastError returns any background storage write error as a string.
func (a *App) GetLastError() string {
	if err := a.storage.LastError(); err != nil {
		return err.Error()
	}
	return ""
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

// emitStateChange fires a Wails event so the frontend can react to server-side
// state mutations (e.g., if a second instance pushes an update in future).
func (a *App) emitStateChange() {
	if a.ctx == nil {
		return
	}
	runtime.EventsEmit(a.ctx, "state:changed", a.state)
}
