package main

import (
	"context"
	"fmt"
	"os"
	"slices"
	"unicode"

	"github.com/Creskendoll/type-buddy/llm"
	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	ctx := context.Background()

	a := app.New()

	w := a.NewWindow("Predictor")

	predictionText := widget.NewLabel("Start typing!")
	w.SetContent(container.NewVBox(predictionText))

	llmClient, err := llm.LLMClient(ctx)
	if err != nil {
		fmt.Println("Error creating LLM client:", err)
		os.Exit(1)
	}

	keyboardEventChannel := make(chan hook.Event)
	mouseEventChannel := make(chan hook.Event)

	// Produce mouse & keyboard events
	go func() {
		eventChain := hook.Start()
		defer hook.End()

		for event := range eventChain {
			if event.Kind == hook.MouseDown {
				mouseEventChannel <- event
			} else if event.Kind == hook.KeyDown {
				keyboardEventChannel <- event
			}
		}
	}()

	// Consume keyboard events
	go func() {
		textBuffer := ""
		prediction := ""
		keysToAccept := []uint16{65507, 65507}
		keysToAcceptState := slices.Clone(keysToAccept)

		for event := range keyboardEventChannel {
			fmt.Println(event)

			if event.Kind == hook.KeyDown {
				keychar := hook.RawcodetoKeychar(event.Rawcode)

				// If the keysToAcceptState contain the keycode, remove it from the state
				if slices.Contains(keysToAcceptState, event.Rawcode) {
					keysToAcceptState = slices.Delete(keysToAcceptState, slices.Index(keysToAcceptState, event.Rawcode), slices.Index(keysToAcceptState, event.Rawcode)+1)
				}

				if !slices.Contains(keysToAccept, event.Rawcode) && len(keysToAcceptState) != len(keysToAccept) {
					fmt.Println("Resetting state")
					keysToAcceptState = slices.Clone(keysToAccept)
				}

				if len(keysToAcceptState) == 0 {
					fmt.Println("Accepting prediction")
					robotgo.TypeStr(prediction)
					keysToAcceptState = slices.Clone(keysToAccept)
				}

				// Backspace
				if event.Rawcode == 65288 && len(textBuffer) > 0 {
					textBuffer = textBuffer[:len(textBuffer)-1]
					continue
				}

				if !unicode.IsPrint(rune(keychar[0])) {
					continue
				}

				textBuffer += keychar

				if len(textBuffer) > 50 {
					textBuffer = ""
				}

				newPrediction, err := llm.Predict(llmClient, ctx, textBuffer)
				if err != nil {
					fmt.Println("Error getting prediction:", err)
					continue
				}

				if newPrediction == "{KO}" {
					continue
				}

				prediction = newPrediction
				fyne.Do(func() {
					predictionText.SetText(prediction)
				})
			}

		}
	}()

	// Consume mouse events
	go func() {
		for mouseEvent := range mouseEventChannel {
			fmt.Println(mouseEvent)
		}
	}()

	w.ShowAndRun()
}
