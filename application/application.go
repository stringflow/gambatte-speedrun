package application

import (
	"github.com/veandco/go-sdl2/sdl"
)

type Application struct {
	EventChannel  chan sdl.Event
	WindowManager *WindowManager
	InputManager  *InputManager
	Audio         *Audio

	GambatteDriver *GambatteDriver
	MainWindow     *MainWindow
	InputDisplay   *InputDisplay
}

func NewApplication(eventChannel chan sdl.Event) (*Application, error) {
	settings := SettingsFromFile("settings.json")

	windowManager := NewWindowManager()
	inputManager := NewInputManager(settings.Inputs, settings.JoystickName)
	audio, err := NewAudio(settings.AudioDeviceName, settings.Volume)
	if err != nil {
		return nil, err
	}

	gambatte := NewGambatte()
	gambatteDriver := &GambatteDriver{
		Gambatte:     gambatte,
		Audio:        audio,
		InputManager: inputManager,
		BufferSize:   settings.BufferSize,
	}

	mainWindowWindow, err := NewWindow("gambatte", 1, 1)
	if err != nil {
		return nil, err
	}
	mainWindow, err := NewMainWindow(mainWindowWindow, gambatte, gambatteDriver, audio, inputManager, windowManager, settings.GameScale, settings.ScaleMode)
	if err != nil {
		return nil, err
	}

	inputDisplayWindow, err := NewWindow("Input Display", KeySize*9, KeySize*8)
	if err != nil {
		return nil, err
	}
	inputDisplay, err := NewInputDisplay(inputDisplayWindow, inputManager, settings.InputDisplayPalette)
	if err != nil {
		return nil, err
	}

	windowManager.SetMainWindow(mainWindow)
	windowManager.AddOptionalWindow(inputDisplay)

	return &Application{
		EventChannel:  eventChannel,
		WindowManager: windowManager,
		InputManager:  inputManager,
		Audio:         audio,

		GambatteDriver: gambatteDriver,
		MainWindow:     mainWindow,
		InputDisplay:   inputDisplay,
	}, nil
}

func (app *Application) Close() {
	NewSettings(app).Save("settings.json")

	app.Audio.Close()
	app.InputManager.Close()
	app.WindowManager.Close()
}

func (app *Application) Update() bool {
	keepRunning := true

	select {
	case event := <-app.EventChannel:
		app.InputManager.ProcessEvent(event)
		keepRunning = app.WindowManager.DispatchEventToAllGuiWindows(event)

		if event.GetType() == sdl.QUIT {
			keepRunning = false
		}
	default:
		app.InputManager.Update()
		app.Draw()
	}

	return keepRunning
}

func (app *Application) Draw() {
	if !app.MainWindow.PopupOpenLastFrame {
		app.GambatteDriver.UpdateEmulator()
	}

	app.WindowManager.DrawAllGuiWindows()
}
