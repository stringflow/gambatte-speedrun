package main

import (
	"time"

	"github.com/stringflow/gambatte-speedrun/application"
	"github.com/veandco/go-sdl2/sdl"
)

func unrecoverableError(err error) {
	sdl.ShowSimpleMessageBox(sdl.MESSAGEBOX_ERROR, "Error", err.Error(), nil)
	panic(err)
}

func main() {
	err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_AUDIO | sdl.INIT_EVENTS | sdl.INIT_JOYSTICK)
	if err != nil {
		unrecoverableError(err)
	}
	defer sdl.Quit()

	// NOTE(stringflow): Allow joystick events to be polled even when the window is not in focus.
	sdl.SetHint(sdl.HINT_JOYSTICK_ALLOW_BACKGROUND_EVENTS, "1")

	eventChannel := make(chan sdl.Event, 64)

	app, err := application.NewApplication(eventChannel)
	if err != nil {
		unrecoverableError(err)
	}
	defer app.Close()

	running := true

	// NOTE(stringflow): The drawing of the emulator is relegated to its own goroutine.
	// This is because polling events may block the goroutine (e.g. when the window is being moved or resized),
	// and we want to keep the emulator running at full speed.
	// SDL events are communicated through the eventChannel for the emulator to process.
	go func() {
		for app.Update() {
		}
		running = false
	}()

	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			eventChannel <- event
		}
		time.Sleep(5 * time.Millisecond)
	}
}
