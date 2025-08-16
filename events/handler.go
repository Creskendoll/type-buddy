package events

import (
	"context"
	"fmt"
	"slices"
	"time"
	"unicode"

	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"

	"github.com/Creskendoll/type-buddy/llm"
	debounce "github.com/Creskendoll/type-buddy/timing"
)

const (
	PredictionEventKind = "prediction"
	CorrectionEventKind = "correction"
	MouseClickEventKind = "mouse_click"
	TextBufferEventKind = "text_buffer"
)

type AppEvent struct {
	Kind    string `json:"kind"`
	Payload string `json:"payload"`

	X int16 `json:"x"`
	Y int16 `json:"y"`
}

func RunEventLoop(llm *llm.LLMClient, ctx context.Context, appEventChannel chan AppEvent) {
	mouseEventChannel := make(chan hook.Event)
	keyboardEventChannel := make(chan hook.Event)

	// Produce mouse & keyboard events
	go func() {
		eventChain := hook.Start()
		defer hook.End()

		for event := range eventChain {
			switch event.Kind {
			case hook.MouseDown:
				mouseEventChannel <- event
			case hook.KeyDown, hook.KeyUp:
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

		predictionDebounced := debounce.New(300 * time.Millisecond)
		correctionDebounced := debounce.New(300 * time.Millisecond)

		for event := range keyboardEventChannel {
			if event.Kind == hook.KeyDown {
				keychar := hook.RawcodetoKeychar(event.Rawcode)

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

				appEventChannel <- AppEvent{
					Kind:    TextBufferEventKind,
					Payload: textBuffer,
				}

				predictionDebounced(func() {
					newPrediction, err := llm.Predict(textBuffer)
					if err != nil {
						fmt.Println("Error getting prediction:", err)
						return
					}

					if newPrediction == "{KO}" {
						return
					}

					appEventChannel <- AppEvent{
						Kind:    PredictionEventKind,
						Payload: newPrediction,
					}
					prediction = newPrediction
				})

				correctionDebounced(func() {
					newCorrected, err := llm.Correct(textBuffer)
					if err != nil {
						fmt.Println("Error getting corrected text:", err)
						return
					}

					if newCorrected == "{KO}" {
						return
					}

					appEventChannel <- AppEvent{
						Kind:    CorrectionEventKind,
						Payload: newCorrected,
					}
				})
			} else if event.Kind == hook.KeyUp {
				// Reset the accept shortcut state
				if len(keysToAcceptState) != len(keysToAccept) && !slices.Contains(keysToAccept, event.Rawcode) {
					keysToAcceptState = slices.Clone(keysToAccept)
				}

				// If the keysToAcceptState contain the keycode, remove it from the state
				if slices.Contains(keysToAcceptState, event.Rawcode) {
					indexToRemove := slices.Index(keysToAcceptState, event.Rawcode)
					keysToAcceptState = slices.Delete(keysToAcceptState, indexToRemove, indexToRemove+1)
				}

				if len(keysToAcceptState) == 0 {
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
			appEventChannel <- AppEvent{
				Kind: MouseClickEventKind,
				X:    mouseEvent.X,
				Y:    mouseEvent.Y,
			}
		}
	}()
}
