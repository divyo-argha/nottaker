package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nottaker/nottaker/core"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx     context.Context
	storage *core.Storage
	state   core.State
	mu      sync.Mutex
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
