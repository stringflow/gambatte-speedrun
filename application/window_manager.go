package application

import (
	"github.com/veandco/go-sdl2/sdl"
)

type GuiWindow interface {
	Draw()
	Close()
	GetWindow() *Window
	GetName() string
}

type WindowManager struct {
	MainWindow      *MainWindow
	OptionalWindows []GuiWindow
}

func NewWindowManager() *WindowManager {
	return &WindowManager{
		OptionalWindows: make([]GuiWindow, 0),
	}
}

func (manager *WindowManager) SetMainWindow(mainWindow *MainWindow) {
	manager.MainWindow = mainWindow
	mainWindow.GetWindow().SetVisible(true)
}

func (manager *WindowManager) AddOptionalWindow(guiWindow GuiWindow) {
	manager.OptionalWindows = append(manager.OptionalWindows, guiWindow)
}

func (manager *WindowManager) ForEachGuiWindow(f func(guiWindow GuiWindow)) {
	f(manager.MainWindow)
	for _, guiWindow := range manager.OptionalWindows {
		f(guiWindow)
	}
}

func (manager *WindowManager) Close() {
	manager.ForEachGuiWindow(func(guiWindow GuiWindow) {
		guiWindow.Close()
	})
}

func (manager *WindowManager) DispatchEventToAllGuiWindows(event sdl.Event) bool {
	manager.ForEachGuiWindow(func(guiWindow GuiWindow) {
		guiWindow.GetWindow().ProcessEvent(event)
	})

	switch e := event.(type) {
	case sdl.WindowEvent:
		if e.Event == sdl.WINDOWEVENT_CLOSE {
			if e.WindowID == manager.MainWindow.GetWindow().Id {
				return false
			}

			manager.ForEachGuiWindow(func(guiWindow GuiWindow) {
				window := guiWindow.GetWindow()
				if e.WindowID == window.Id {
					window.SetVisible(false)
				}
			})
		}
	}

	return true
}

func (manager *WindowManager) DrawAllGuiWindows() {
	manager.ForEachGuiWindow(func(guiWindow GuiWindow) {
		window := guiWindow.GetWindow()
		if !window.Visible() {
			return
		}

		window.Renderer.NewFrame()
		guiWindow.Draw()
		window.Renderer.EndFrame()
	})
}
