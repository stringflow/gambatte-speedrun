package application

import (
	imgui "github.com/AllenDang/cimgui-go"
	"github.com/veandco/go-sdl2/sdl"
)

const BaseFontSize = 13.0

type Window struct {
	Id        uint32
	SdlWindow *sdl.Window
	Renderer  *Renderer
}

func NewWindow(title string, width int, height int) (*Window, error) {
	window, err := sdl.CreateWindow(title, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, int32(width), int32(height), sdl.WINDOW_HIDDEN)
	if err != nil {
		return nil, err
	}

	windowId, err := window.GetID()
	if err != nil {
		return nil, err
	}

	renderer, err := NewRenderer(window)
	if err != nil {
		return nil, err
	}

	return &Window{
		Id:        windowId,
		SdlWindow: window,
		Renderer:  renderer,
	}, nil
}

func (window *Window) Close() {
	window.Renderer.Close()
	window.SdlWindow.Destroy()
}

func (window *Window) SetVisible(visible bool) {
	if visible {
		window.SdlWindow.Show()
	} else {
		window.SdlWindow.Hide()
	}
}

func (window *Window) Visible() bool {
	return (window.SdlWindow.GetFlags() & sdl.WINDOW_SHOWN) > 0
}

func (window *Window) ProcessEvent(event sdl.Event) {
	window.Renderer.ProcessEvent(event)
}

type Renderer struct {
	SdlRenderer  *sdl.Renderer
	ImGuiContext *imgui.Context
	ImGuiBackend *ImGuiBackend
}

func NewRenderer(sdlWindow *sdl.Window) (*Renderer, error) {
	sdlRenderer, err := sdl.CreateRenderer(sdlWindow, -1, sdl.RENDERER_SOFTWARE)
	if err != nil {
		return nil, err
	}

	imguiContext := imgui.CreateContext()
	imgui.SetCurrentContext(imguiContext)
	imgui.CurrentIO().SetIniFilename("")

	imguiBackend, err := NewImGuiBackend(sdlWindow, sdlRenderer)
	if err != nil {
		return nil, err
	}

	return &Renderer{
		SdlRenderer:  sdlRenderer,
		ImGuiContext: imguiContext,
		ImGuiBackend: imguiBackend,
	}, nil
}

func (renderer *Renderer) Close() {
	imgui.SetCurrentContext(renderer.ImGuiContext)
	renderer.ImGuiBackend.Close()
	renderer.ImGuiContext.Destroy()
	renderer.SdlRenderer.Destroy()
}

func (renderer *Renderer) ProcessEvent(event sdl.Event) {
	imgui.SetCurrentContext(renderer.ImGuiContext)
	renderer.ImGuiBackend.ProcessEvent(event)
}

func (renderer *Renderer) NewFrame() {
	imgui.SetCurrentContext(renderer.ImGuiContext)
	renderer.ImGuiBackend.NewFrame()
	imgui.NewFrame()
}

func (renderer *Renderer) EndFrame() {
	imgui.SetCurrentContext(renderer.ImGuiContext)
	imgui.Render()
	renderer.SdlRenderer.Clear()
	renderer.ImGuiBackend.Render()
	renderer.SdlRenderer.Present()
}
