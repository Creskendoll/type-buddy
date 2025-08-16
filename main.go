package main

import (
	"context"
	"embed"
	"fmt"
	"os"

	"github.com/Creskendoll/type-buddy/events"
	"github.com/Creskendoll/type-buddy/llm"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	ctx := context.Background()

	// Create an instance of the app structure
	app := NewApp()

	llmClient, err := llm.New(ctx, "")
	if err != nil {
		fmt.Println("Error creating LLM client:", err)
		os.Exit(1)
	}

	appEventChannel := make(chan events.AppEvent)
	events.RunEventLoop(llmClient, ctx, appEventChannel)

	go func() {
		for event := range appEventChannel {
			switch event.Kind {
			case events.PredictionEventKind:
				runtime.EventsEmit(app.ctx, SetPredictionTextEvent, event.Payload)
			case events.CorrectionEventKind:
				runtime.EventsEmit(app.ctx, SetCorrectedTextEvent, event.Payload)
			case events.TextBufferEventKind:
				runtime.EventsEmit(app.ctx, SetBufferTextEvent, event.Payload)
			case events.MouseClickEventKind:
				fmt.Println("Mouse clicked at:", event.X, event.Y)
				runtime.WindowSetPosition(app.ctx, int(event.X), int(event.Y))
			}
		}
	}()

	// Create application with options
	err = wails.Run(&options.App{
		Title:       "type-buddy-ui",
		Width:       500,
		Height:      300,
		AlwaysOnTop: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Bind:      []any{app},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
