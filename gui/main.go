package main

import (
	"context"
	"fmt"
	"os"

	"github.com/nottaker/octonote/core"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

func main() {
	storage, err := core.NewStorage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "octoNote gui: storage init: %v\n", err)
		os.Exit(1)
	}

	app := NewApp(storage)

	err = wails.Run(&options.App{
		Title:     "octoNote",
		Width:     1024,
		Height:    720,
		MinWidth:  720,
		MinHeight: 480,

		Frameless:        true,
		BackgroundColour: &options.RGBA{R: 15, G: 15, B: 15, A: 255},

		AssetServer: &assetserver.Options{
			Assets: assets,
		},

		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               "io.octonote.app.singleinstance",
			OnSecondInstanceLaunch: app.onSecondInstanceLaunch,
		},

		OnStartup:  app.startup,
		OnShutdown: app.shutdown,

		Bind: []interface{}{app},

		Mac: &mac.Options{
			TitleBar:             mac.TitleBarHiddenInset(),
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			About: &mac.AboutInfo{
				Title:   "octoNote",
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
		fmt.Fprintf(os.Stderr, "octoNote gui: %v\n", err)
		os.Exit(1)
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	st, err := a.storage.Load()
	if err != nil {
		st = core.State{
			Version:     1,
			ActiveIndex: 0,
			Tabs:        []core.Tab{core.NewTab("scratch")},
		}
	}
	a.state = st
}

func (a *App) shutdown(_ context.Context) {
	a.storage.Save(a.state)
	a.storage.Close()
}

func (a *App) onSecondInstanceLaunch(_ options.SecondInstanceData) {
}
