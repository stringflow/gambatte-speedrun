package application

import (
	"encoding/json"
	"os"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

type Settings struct {
	AudioDeviceName string
	Volume          int32

	Inputs       InputArray
	JoystickName string

	BufferSize time.Duration
	GameScale  int32
	ScaleMode  sdl.ScaleMode

	InputDisplayPalette string
}

func NewSettings(app *Application) *Settings {
	return &Settings{
		AudioDeviceName:     app.Audio.OpenedDeviceName,
		Volume:              app.Audio.Volume,
		Inputs:              app.InputManager.Inputs,
		JoystickName:        app.InputManager.OpenedJoystickName,
		BufferSize:          app.GambatteDriver.BufferSize,
		GameScale:           app.MainWindow.Scale,
		ScaleMode:           app.MainWindow.GambatteSource.Backbuffer.ScaleMode(),
		InputDisplayPalette: app.InputDisplay.SelectedPalette,
	}
}

func DefaultSettings() *Settings {
	return &Settings{
		AudioDeviceName: SystemDefaultAudioDevice,
		Volume:          1,
		Inputs: InputArray{
			A:                 NewKeyInput(sdl.K_z, sdl.KMOD_NONE),
			B:                 NewKeyInput(sdl.K_x, sdl.KMOD_NONE),
			Select:            NewKeyInput(sdl.K_BACKSPACE, sdl.KMOD_NONE),
			Start:             NewKeyInput(sdl.K_RETURN, sdl.KMOD_NONE),
			Right:             NewKeyInput(sdl.K_RIGHT, sdl.KMOD_NONE),
			Left:              NewKeyInput(sdl.K_LEFT, sdl.KMOD_NONE),
			Up:                NewKeyInput(sdl.K_UP, sdl.KMOD_NONE),
			Down:              NewKeyInput(sdl.K_DOWN, sdl.KMOD_NONE),
			Reset:             NewKeyInput(sdl.K_r, sdl.KMOD_NONE),
			SaveState:         NewKeyInput(sdl.K_F1, sdl.KMOD_NONE),
			LoadState:         NewKeyInput(sdl.K_F2, sdl.KMOD_NONE),
			PreviousStateSlot: NewKeyInput(sdl.K_F3, sdl.KMOD_NONE),
			NextStateSlot:     NewKeyInput(sdl.K_F4, sdl.KMOD_NONE),
		},
		JoystickName:        "",
		BufferSize:          68 * time.Millisecond,
		GameScale:           3,
		ScaleMode:           sdl.ScaleModeNearest,
		InputDisplayPalette: "Monochrome",
	}
}

func SettingsFromFile(path string) *Settings {
	settings := DefaultSettings()

	jsonData, err := os.ReadFile(path)
	if err != nil {
		return settings
	}

	err = json.Unmarshal(jsonData, &settings)
	if err != nil {
		return settings
	}

	return settings
}

func (settings Settings) Save(path string) {
	jsonData, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return
	}

	os.WriteFile(path, jsonData, 0644)
}
