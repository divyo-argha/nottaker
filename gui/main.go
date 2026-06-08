package main

import (
	"context"
	"fmt"
	"os"

	"github.com/nottaker/nottaker/core"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

func main() {
	// Initialise shared storage backend.
	storage, err := core.NewStorage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "nottaker gui: storage init: %v\n", err)
		os.Exit(1)
	}

	app := NewApp(storage)

	err = wails.Run(&options.App{
		Title:  "nottaker",
		Width:  1024,
		Height: 720,
		MinWidth:  720,
		MinHeight: 480,

		// Frameless, dark, translucent — matches TUI aesthetic.
		Frameless:        true,
		BackgroundColour: &options.RGBA{R: 15, G: 15, B: 15, A: 255},

		AssetServer: &assetserver.Options{
			Assets: assets, // embedded in assets.go
		},

		// Single-instance: bring existing window to front.
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               "io.nottaker.app.singleinstance",
			OnSecondInstanceLaunch: app.onSecondInstanceLaunch,
		},

		OnStartup:  app.startup,
		OnShutdown: app.shutdown,

		Bind: []interface{}{app},

		// Platform-specific window chrome.
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarHiddenInset(),
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			About: &mac.AboutInfo{
				Title:   "nottaker",
				Message: "Lightweight multi-tab scratchpad.\n\nBuilt with Wails & Go.",
			},
		},
		Windows: &windows.Options{
			WebviewIsTransparent:              false,
			WindowIsTranslucent:               false,
			DisableFramelessWindowDecorations: false,
		},
		Linux: &linux.Options{
			WindowIsTranslucent: false,
		},
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "nottaker gui: %v\n", err)
		os.Exit(1)
	}
}

// startup is called once Wails has finished launching.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	st, err := a.storage.Load()
	if err != nil {
		// Non-fatal: fall back to a default single-tab state.
		st = core.State{
			Version:     1,
			ActiveIndex: 0,
			Tabs:        []core.Tab{core.NewTab("scratch")},
		}
	}
	a.state = st
}

// shutdown is called when the app is about to close.
func (a *App) shutdown(_ context.Context) {
	a.storage.Save(a.state)
	a.storage.Close()
}

// onSecondInstanceLaunch is called when a second instance tries to start.
func (a *App) onSecondInstanceLaunch(_ options.SecondInstanceData) {
	// Bring existing window to front (runtime.WindowShow would go here in a full impl).
}
