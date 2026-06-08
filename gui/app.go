package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nottaker/octonote/core"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// shareSession holds the cancel function for an in-flight share operation.
type shareSession struct {
	cancel context.CancelFunc
}

type App struct {
	ctx     context.Context
	storage *core.Storage
	state   core.State
	mu      sync.Mutex

	// share session — only one active at a time
	shareMu  sync.Mutex
	shareSes *shareSession
}

func NewApp(storage *core.Storage) *App {
	return &App{storage: storage}
}

func (a *App) GetState() core.State {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.state
}

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

func (a *App) CloseTab(index int) core.State {
	a.mu.Lock()
	defer a.mu.Unlock()
	if index < 0 || index >= len(a.state.Tabs) {
		return a.state
	}
	if len(a.state.Tabs) == 1 {
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

func (a *App) GetStorageDir() string {
	return a.storage.Dir()
}

func (a *App) GetLastError() string {
	if err := a.storage.LastError(); err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) emitStateChange() {
	if a.ctx == nil {
		return
	}
	runtime.EventsEmit(a.ctx, "state:changed", a.state)
}

// ── Share feature ────────────────────────────────────────────────────────────

// ShareSend opens a Magic Wormhole for the currently active tab.
// senderLabel is an optional name the sender provides (e.g. "Alice") so the
// receiver can see who shared the content. Pass an empty string to omit.
// It immediately emits "share:code" with the generated code, then waits
// for the peer to connect. On success it emits "share:done", on error
// it emits "share:error".
func (a *App) ShareSend(senderLabel string) {
	a.mu.Lock()
	idx := a.state.ActiveIndex
	var tab core.Tab
	if idx >= 0 && idx < len(a.state.Tabs) {
		tab = a.state.Tabs[idx]
	}
	a.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	a.shareMu.Lock()
	if a.shareSes != nil {
		a.shareSes.cancel() // cancel any previous session
	}
	a.shareSes = &shareSession{cancel: cancel}
	a.shareMu.Unlock()

	go func() {
		defer cancel()
		code, wait, err := core.ShareSend(ctx, tab, senderLabel)
		if err != nil {
			runtime.EventsEmit(a.ctx, "share:error", err.Error())
			return
		}
		runtime.EventsEmit(a.ctx, "share:code", code)
		if err := wait(); err != nil {
			// Context cancelled = user hit cancel — emit nothing.
			if ctx.Err() == nil {
				runtime.EventsEmit(a.ctx, "share:error", err.Error())
			}
			return
		}
		runtime.EventsEmit(a.ctx, "share:done", nil)
	}()
}

// ShareReceive connects to a wormhole using the user-supplied code.
// On success it imports the received content as a new tab and emits
// "share:received" with the new tab title. On error it emits "share:error".
func (a *App) ShareReceive(code string) {
	if code == "" {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.shareMu.Lock()
	if a.shareSes != nil {
		a.shareSes.cancel()
	}
	a.shareSes = &shareSession{cancel: cancel}
	a.shareMu.Unlock()

	go func() {
		defer cancel()
		result, err := core.ShareReceive(ctx, code)
		if err != nil {
			if ctx.Err() == nil {
				runtime.EventsEmit(a.ctx, "share:error", err.Error())
			}
			return
		}
		a.mu.Lock()
		newTab := core.Tab{
			ID:        fmt.Sprintf("%x", time.Now().UnixNano()),
			Title:     result.TabTitle,
			Body:      result.Body,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		a.state.Tabs = append(a.state.Tabs, newTab)
		a.state.ActiveIndex = len(a.state.Tabs) - 1
		a.storage.Save(a.state)
		a.mu.Unlock()
		runtime.EventsEmit(a.ctx, "share:received", map[string]interface{}{
			"title": result.TabTitle,
			"state": a.state,
		})
	}()
}

// ShareCancel aborts any in-flight share or receive operation.
func (a *App) ShareCancel() {
	a.shareMu.Lock()
	defer a.shareMu.Unlock()
	if a.shareSes != nil {
		a.shareSes.cancel()
		a.shareSes = nil
	}
}
