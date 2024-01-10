package application

import (
	"fmt"

	imgui "github.com/AllenDang/cimgui-go"
	"github.com/veandco/go-sdl2/sdl"
)

type MainWindow struct {
	Window             *Window
	GambatteSource     *GambatteSource
	GambatteDriver     *GambatteDriver
	SettingsModal      *SettingsModal
	Scale              int32
	PopupOpenLastFrame bool
}

func NewMainWindow(window *Window, gambatte *Gambatte, gambatteDriver *GambatteDriver, audio *Audio, inputManager *InputManager, windowManager *WindowManager, scale int32, scaleMode sdl.ScaleMode) (*MainWindow, error) {
	gambatteSource, err := NewGambatteSource(window.Renderer, gambatte, inputManager, windowManager, scaleMode)
	if err != nil {
		return nil, err
	}

	settingsModal := NewSettingsModal(inputManager, audio, windowManager, gambatteSource, gambatteDriver)

	mainWindow := &MainWindow{
		Window:         window,
		GambatteSource: gambatteSource,
		SettingsModal:  settingsModal,
	}
	mainWindow.SetScale(scale)

	return mainWindow, nil
}

func (mw *MainWindow) Draw() {
	mw.GambatteSource.Draw()
	mw.SettingsModal.Draw()

	// TODO(stringflow): Massive hack, but it works
	mw.PopupOpenLastFrame = fmt.Sprintf("%#v", imgui.InternalTopMostAndVisiblePopupModal()) != fmt.Sprintf("%#v", &imgui.Window{})
}

func (mw *MainWindow) SetScale(scale int32) {
	mw.Window.SdlWindow.SetSize(scale*GbVideoWidth, scale*GbVideoHeight+19)
	mw.Scale = scale
}

func (mw MainWindow) Close() {
	mw.GambatteSource.Close()
}

func (mw MainWindow) GetWindow() *Window {
	return mw.Window
}

func (mw MainWindow) GetName() string {
	return "Main Window"
}
