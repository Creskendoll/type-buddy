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

	llmClient, err := llm.LLMClient(ctx)
	if err != nil {
		fmt.Println("Error creating LLM client:", err)
		os.Exit(1)
	}

	setPredictionText := func(text string) {
		runtime.EventsEmit(app.ctx, "setPredictionText", text)
	}

	setBufferText := func(text string) {
		runtime.EventsEmit(app.ctx, "setBufferText", text)
	}

	events.RunEventLoop(llmClient, ctx, setPredictionText, setBufferText)

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
