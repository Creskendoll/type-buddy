package main

import (
	"context"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

const (
	SetPredictionTextEvent = "setPredictionText"
	SetBufferTextEvent     = "setBufferText"
	SetCorrectedTextEvent  = "setCorrectedText"
)
