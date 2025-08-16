package events

import (
	"context"
	"fmt"
	"slices"
	"time"
	"unicode"

	"github.com/go-vgo/robotgo"
	"github.com/ollama/ollama/api"
	hook "github.com/robotn/gohook"

	"github.com/Creskendoll/type-buddy/llm"
	debounce "github.com/Creskendoll/type-buddy/timing"
)

func RunEventLoop(llmClient *api.Client, ctx context.Context, setPredictionText func(string), setBufferText func(string)) {
	mouseEventChannel := make(chan hook.Event)
	keyboardEventChannel := make(chan hook.Event)

	// Produce mouse & keyboard events
	go func() {
		eventChain := hook.Start()
		defer hook.End()

		for event := range eventChain {
			if event.Kind == hook.MouseDown {
				mouseEventChannel <- event
			} else if event.Kind == hook.KeyDown || event.Kind == hook.KeyUp {
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

		debounced := debounce.New(300 * time.Millisecond)

		for event := range keyboardEventChannel {
			fmt.Println(event)

			if event.Kind == hook.KeyDown {
				keychar := hook.RawcodetoKeychar(event.Rawcode)

				// Reset the accept shortcut state
				if len(keysToAcceptState) != len(keysToAccept) && !slices.Contains(keysToAccept, event.Rawcode) {
					keysToAcceptState = slices.Clone(keysToAccept)
				}

				// Backspace
				if event.Rawcode == 65288 && len(textBuffer) > 0 {
					textBuffer = textBuffer[:len(textBuffer)-1]
				} else if unicode.IsPrint(rune(keychar[0])) {
					textBuffer += keychar
				} else {
					// A modifier key was pressed, do not update the text buffer
					continue
				}

				if len(textBuffer) > 100 {
					textBuffer = ""
				}

				setBufferText(textBuffer)

				debounced(func() {
					newPrediction, err := llm.Predict(llmClient, ctx, textBuffer)
					if err != nil {
						fmt.Println("Error getting prediction:", err)
						return
					}

					if newPrediction == "{KO}" {
						return
					}

					setPredictionText(newPrediction)
					prediction = newPrediction
				})
			} else if event.Kind == hook.KeyUp {
				// If the keysToAcceptState contain the keycode, remove it from the state
				if slices.Contains(keysToAcceptState, event.Rawcode) {
					indexToRemove := slices.Index(keysToAcceptState, event.Rawcode)
					keysToAcceptState = slices.Delete(keysToAcceptState, indexToRemove, indexToRemove+1)
				}

				if len(keysToAcceptState) == 0 {
					fmt.Println("Accepting prediction:", prediction)
					robotgo.TypeStrDelay(prediction, 1000)
					keysToAcceptState = slices.Clone(keysToAccept)
				}
			}
		}
	}()

	// Consume mouse events
	go func() {
		for mouseEvent := range mouseEventChannel {
			fmt.Println(mouseEvent)
		}
	}()
}
