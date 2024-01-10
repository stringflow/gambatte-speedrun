package application

import (
	"math"
	"slices"
	"unsafe"

	imgui "github.com/AllenDang/cimgui-go"
	"github.com/veandco/go-sdl2/sdl"
)

type ImGuiBackend struct {
	WindowId               uint32
	SdlWindow              *sdl.Window
	SdlRenderer            *sdl.Renderer
	SdlTexture             *sdl.Texture
	SdlCursors             map[imgui.MouseCursor]*sdl.Cursor
	PerformanceFrequency   uint64
	MouseCanUseGlobalState bool

	Time                   uint64
	PendingMouseLeaveFrame int
	LastCursor             *sdl.Cursor
}

func NewImGuiBackend(window *sdl.Window, renderer *sdl.Renderer) (*ImGuiBackend, error) {
	windowId, err := window.GetID()
	if err != nil {
		return nil, err
	}

	io := imgui.CurrentIO()
	io.SetBackendRendererName("sdlrenderer")
	io.SetBackendFlags(io.BackendFlags() | imgui.BackendFlagsRendererHasVtxOffset | imgui.BackendFlagsHasMouseCursors)
	sdl.SetHint(sdl.HINT_MOUSE_AUTO_CAPTURE, "0")

	pixels, width, height, _ := io.Fonts().GetTextureDataAsRGBA32()

	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STATIC, width, height)
	if err != nil {
		return nil, err
	}

	err = texture.Update(nil, pixels, int(width*4))
	if err != nil {
		return nil, err
	}

	texture.SetBlendMode(sdl.BLENDMODE_BLEND)

	io.Fonts().SetTexID(imgui.TextureID(texture))

	cursors := make(map[imgui.MouseCursor]*sdl.Cursor)
	cursors[imgui.MouseCursorArrow] = sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_ARROW)
	cursors[imgui.MouseCursorTextInput] = sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_IBEAM)
	cursors[imgui.MouseCursorResizeAll] = sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_SIZEALL)
	cursors[imgui.MouseCursorResizeNS] = sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_SIZENS)
	cursors[imgui.MouseCursorResizeEW] = sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_SIZEWE)
	cursors[imgui.MouseCursorResizeNESW] = sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_SIZENESW)
	cursors[imgui.MouseCursorResizeNWSE] = sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_SIZENWSE)
	cursors[imgui.MouseCursorHand] = sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_HAND)
	cursors[imgui.MouseCursorNotAllowed] = sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_NO)

	sdlVideoDriver, err := sdl.GetCurrentVideoDriver()
	if err != nil {
		return nil, err
	}

	backend := &ImGuiBackend{
		WindowId:               windowId,
		SdlWindow:              window,
		SdlRenderer:            renderer,
		SdlTexture:             texture,
		SdlCursors:             cursors,
		PerformanceFrequency:   sdl.GetPerformanceFrequency(),
		MouseCanUseGlobalState: slices.Contains([]string{"windows", "cocoa", "x11", "DIVE", "VMAN"}, sdlVideoDriver),
	}

	io.SetClipboardHandler(backend)

	return backend, nil
}

func (backend *ImGuiBackend) Close() {
	for _, cursor := range backend.SdlCursors {
		sdl.FreeCursor(cursor)
	}

	backend.SdlTexture.Destroy()
}

func (backend *ImGuiBackend) NewFrame() {
	backend.updateDisplay()
	backend.updateTime()
	backend.updateMouse()
	backend.updateMouseCursor()
}

func (backend *ImGuiBackend) updateDisplay() {
	io := imgui.CurrentIO()

	windowWidth, windowHeight := backend.SdlWindow.GetSize()

	if (backend.SdlWindow.GetFlags() & sdl.WINDOW_MINIMIZED) != 0 {
		windowWidth = 0
		windowHeight = 0
	}

	framebufferWidth, framebufferHeight, err := backend.SdlRenderer.GetOutputSize()
	if err == nil {
		framebufferWidth = windowWidth
		framebufferHeight = windowHeight
	}

	io.SetDisplaySize(imgui.Vec2{X: float32(windowWidth), Y: float32(windowHeight)})
	if windowWidth > 0 && windowHeight > 0 {
		io.SetDisplayFramebufferScale(imgui.Vec2{X: float32(framebufferWidth) / float32(windowWidth), Y: float32(framebufferHeight) / float32(windowHeight)})
	}
}

func (backend *ImGuiBackend) updateTime() {
	io := imgui.CurrentIO()

	currentTime := sdl.GetPerformanceCounter()
	delta := float64(currentTime-backend.Time) / float64(backend.PerformanceFrequency)
	if backend.Time == 0 {
		delta = 1.0 / 60.0
	}

	io.SetDeltaTime(float32(delta))
	backend.Time = currentTime
}

func (backend *ImGuiBackend) updateMouse() {
	io := imgui.CurrentIO()

	_, _, mouseButtonsDown := sdl.GetMouseState()
	areNoButtonsDown := mouseButtonsDown != sdl.ButtonMask(0)

	sdl.CaptureMouse(mouseButtonsDown.Has(sdl.ButtonLeft))

	if backend.PendingMouseLeaveFrame != 0 && backend.PendingMouseLeaveFrame >= int(imgui.FrameCount()) && areNoButtonsDown {
		io.AddMousePosEvent(-math.MaxFloat32, -math.MaxFloat32)
		backend.PendingMouseLeaveFrame = 0
	}

	window := sdl.GetKeyboardFocus()
	is_app_focused := window == backend.SdlWindow

	if is_app_focused {
		io := imgui.CurrentIO()

		if io.WantSetMousePos() {
			window.WarpMouseInWindow(int32(io.MousePos().X), int32(io.MousePos().Y))
		}

		if backend.MouseCanUseGlobalState && areNoButtonsDown {
			global_x, global_y, _ := sdl.GetGlobalMouseState()
			window_x, window_y := window.GetPosition()
			io.AddMousePosEvent(float32(global_x-window_x), float32(global_y-window_y))
		}
	}
}

func (backend *ImGuiBackend) updateMouseCursor() {
	io := imgui.CurrentIO()
	if (io.ConfigFlags() & imgui.ConfigFlagsNoMouseCursorChange) != 0 {
		return
	}

	imguiCursor := imgui.CurrentMouseCursor()
	if io.MouseDrawCursor() || imguiCursor == imgui.MouseCursorNone {
		sdl.ShowCursor(sdl.DISABLE)
	} else {
		sdlCursor, ok := backend.SdlCursors[imguiCursor]
		if !ok {
			sdlCursor = backend.SdlCursors[imgui.MouseCursorArrow]
		}

		if backend.LastCursor != sdlCursor {
			backend.LastCursor = sdlCursor
			sdl.SetCursor(sdlCursor)
		}

		sdl.ShowCursor(sdl.ENABLE)
	}
}

func (backend *ImGuiBackend) ProcessEvent(event sdl.Event) {
	io := imgui.CurrentIO()

	switch e := event.(type) {
	case sdl.MouseMotionEvent:
		if e.WindowID == backend.WindowId {
			io.AddMousePosEvent(float32(e.X), float32(e.Y))
		}
	case sdl.MouseWheelEvent:
		io.AddMouseWheelEvent(-e.PreciseX, e.PreciseY)
	case sdl.MouseButtonEvent:
		if e.WindowID == backend.WindowId {
			mouseButton, ok := sdlButtonToImgui[e.Button]
			if !ok {
				return
			}

			io.AddMouseButtonEvent(mouseButton, e.State == sdl.PRESSED)
		}
	case sdl.TextInputEvent:
		if e.WindowID == backend.WindowId {
			io.AddInputCharactersUTF8(e.Text)
		}
	case sdl.KeyboardEvent:
		if e.WindowID == backend.WindowId {
			io.AddKeyEvent(imgui.ModCtrl, (e.Keysym.Mod&uint16(sdl.KMOD_CTRL)) != 0)
			io.AddKeyEvent(imgui.ModShift, (e.Keysym.Mod&uint16(sdl.KMOD_SHIFT)) != 0)
			io.AddKeyEvent(imgui.ModAlt, (e.Keysym.Mod&uint16(sdl.KMOD_ALT)) != 0)
			io.AddKeyEvent(imgui.ModSuper, (e.Keysym.Mod&uint16(sdl.KMOD_GUI)) != 0)

			key, ok := sdlKeycodeToImgui[e.Keysym.Sym]
			if !ok {
				return
			}

			io.AddKeyEvent(key, e.State == sdl.PRESSED)
			io.SetKeyEventNativeData(key, int32(e.Keysym.Sym), int32(e.Keysym.Scancode))
		}
	case sdl.WindowEvent:
		if e.WindowID == backend.WindowId {
			switch e.Event {
			case sdl.WINDOWEVENT_ENTER:
				backend.PendingMouseLeaveFrame = 0
			case sdl.WINDOWEVENT_LEAVE:
				backend.PendingMouseLeaveFrame = int(imgui.FrameCount()) + 1
			case sdl.WINDOWEVENT_FOCUS_GAINED:
				io.AddFocusEvent(true)
			case sdl.WINDOWEVENT_FOCUS_LOST:
				io.AddFocusEvent(false)
			}
		}
	}
}

func (backend *ImGuiBackend) Render() {
	drawData := imgui.CurrentDrawData()

	renderScaleX, renderScaleY := backend.SdlRenderer.GetScale()

	if renderScaleX == 1.0 {
		renderScaleX = drawData.FramebufferScale().X
	}

	if renderScaleY == 1.0 {
		renderScaleY = drawData.FramebufferScale().Y
	}

	renderScale := imgui.Vec2{X: renderScaleX, Y: renderScaleY}
	framebufferWidth := int32(drawData.DisplaySize().X * renderScale.X)
	framebufferHeight := int32(drawData.DisplaySize().Y * renderScale.Y)

	if framebufferWidth == 0 || framebufferHeight == 0 {
		return
	}

	clipOff := drawData.DisplayPos()
	clipScale := renderScale

	vertexSize := int(unsafe.Sizeof(imgui.DrawVert{}))
	indexSize := int(unsafe.Sizeof(imgui.DrawIdx(0)))

	backend.SdlRenderer.SetViewport(nil)
	backend.SdlRenderer.SetClipRect(nil)

	for _, cmdList := range drawData.CommandLists() {
		vertexBuffer, vertexBufferSize := cmdList.GetVertexBuffer()
		indexBuffer, _ := cmdList.GetIndexBuffer()

		vertexCount := vertexBufferSize / vertexSize

		for _, cmd := range cmdList.Commands() {
			if cmd.HasUserCallback() {
				cmd.CallUserCallback(cmdList)
			} else {
				clipMin := imgui.Vec2{
					X: (cmd.ClipRect().X - clipOff.X) * clipScale.X, Y: (cmd.ClipRect().Y - clipOff.Y) * clipScale.Y,
				}

				clipMax := imgui.Vec2{
					X: (cmd.ClipRect().Z - clipOff.X) * clipScale.X, Y: (cmd.ClipRect().W - clipOff.Y) * clipScale.Y,
				}

				if (clipMax.X <= clipMin.X) || (clipMax.Y <= clipMin.Y) {
					continue
				}

				rect := &sdl.Rect{
					X: int32(clipMin.X),
					Y: int32(clipMin.Y),
					W: int32(clipMax.X - clipMin.X),
					H: int32(clipMax.Y - clipMin.Y),
				}
				backend.SdlRenderer.SetClipRect(rect)

				xy := unsafe.Add(vertexBuffer, int(cmd.VtxOffset())*vertexSize)
				uv := unsafe.Add(xy, 8)
				color := unsafe.Add(xy, 16)
				idx := unsafe.Add(indexBuffer, int(cmd.IdxOffset())*indexSize)

				backend.SdlRenderer.RenderGeometryRaw(
					(*sdl.Texture)(cmd.TexID()),
					(*float32)(xy), vertexSize,
					(*sdl.Color)(color), vertexSize,
					(*float32)(uv), vertexSize,
					vertexCount-int(cmd.VtxOffset()),
					idx, int(cmd.ElemCount()), int(indexSize))
			}
		}
	}
}

func (backend *ImGuiBackend) GetClipboard() string {
	text, err := sdl.GetClipboardText()
	if err != nil {
		return ""
	} else {
		return text
	}
}

func (backend *ImGuiBackend) SetClipboard(s string) {
	sdl.SetClipboardText(s)
}

var (
	sdlButtonToImgui = map[sdl.Button]int32{
		sdl.ButtonLeft:   0,
		sdl.ButtonRight:  1,
		sdl.ButtonMiddle: 2,
		sdl.ButtonX1:     3,
		sdl.ButtonX2:     4,
	}

	sdlKeycodeToImgui = map[sdl.Keycode]imgui.Key{
		sdl.K_TAB:          imgui.KeyTab,
		sdl.K_LEFT:         imgui.KeyLeftArrow,
		sdl.K_RIGHT:        imgui.KeyRightArrow,
		sdl.K_UP:           imgui.KeyUpArrow,
		sdl.K_DOWN:         imgui.KeyDownArrow,
		sdl.K_PAGEUP:       imgui.KeyPageUp,
		sdl.K_PAGEDOWN:     imgui.KeyPageDown,
		sdl.K_HOME:         imgui.KeyHome,
		sdl.K_END:          imgui.KeyEnd,
		sdl.K_INSERT:       imgui.KeyInsert,
		sdl.K_DELETE:       imgui.KeyDelete,
		sdl.K_BACKSPACE:    imgui.KeyBackspace,
		sdl.K_SPACE:        imgui.KeySpace,
		sdl.K_RETURN:       imgui.KeyEnter,
		sdl.K_ESCAPE:       imgui.KeyEscape,
		sdl.K_QUOTE:        imgui.KeyApostrophe,
		sdl.K_COMMA:        imgui.KeyComma,
		sdl.K_MINUS:        imgui.KeyMinus,
		sdl.K_PERIOD:       imgui.KeyPeriod,
		sdl.K_SLASH:        imgui.KeySlash,
		sdl.K_SEMICOLON:    imgui.KeySemicolon,
		sdl.K_EQUALS:       imgui.KeyEqual,
		sdl.K_LEFTBRACKET:  imgui.KeyLeftBracket,
		sdl.K_BACKSLASH:    imgui.KeyBackslash,
		sdl.K_RIGHTBRACKET: imgui.KeyRightBracket,
		sdl.K_BACKQUOTE:    imgui.KeyGraveAccent,
		sdl.K_CAPSLOCK:     imgui.KeyCapsLock,
		sdl.K_SCROLLLOCK:   imgui.KeyScrollLock,
		sdl.K_NUMLOCKCLEAR: imgui.KeyNumLock,
		sdl.K_PRINTSCREEN:  imgui.KeyPrintScreen,
		sdl.K_PAUSE:        imgui.KeyPause,
		sdl.K_KP_0:         imgui.KeyKeypad0,
		sdl.K_KP_1:         imgui.KeyKeypad1,
		sdl.K_KP_2:         imgui.KeyKeypad2,
		sdl.K_KP_3:         imgui.KeyKeypad3,
		sdl.K_KP_4:         imgui.KeyKeypad4,
		sdl.K_KP_5:         imgui.KeyKeypad5,
		sdl.K_KP_6:         imgui.KeyKeypad6,
		sdl.K_KP_7:         imgui.KeyKeypad7,
		sdl.K_KP_8:         imgui.KeyKeypad8,
		sdl.K_KP_9:         imgui.KeyKeypad9,
		sdl.K_KP_PERIOD:    imgui.KeyKeypadDecimal,
		sdl.K_KP_DIVIDE:    imgui.KeyKeypadDivide,
		sdl.K_KP_MULTIPLY:  imgui.KeyKeypadMultiply,
		sdl.K_KP_MINUS:     imgui.KeyKeypadSubtract,
		sdl.K_KP_PLUS:      imgui.KeyKeypadAdd,
		sdl.K_KP_ENTER:     imgui.KeyKeypadEnter,
		sdl.K_KP_EQUALS:    imgui.KeyKeypadEqual,
		sdl.K_LCTRL:        imgui.KeyLeftCtrl,
		sdl.K_LSHIFT:       imgui.KeyLeftShift,
		sdl.K_LALT:         imgui.KeyLeftAlt,
		sdl.K_LGUI:         imgui.KeyLeftSuper,
		sdl.K_RCTRL:        imgui.KeyRightCtrl,
		sdl.K_RSHIFT:       imgui.KeyRightShift,
		sdl.K_RALT:         imgui.KeyRightAlt,
		sdl.K_RGUI:         imgui.KeyRightSuper,
		sdl.K_APPLICATION:  imgui.KeyMenu,
		sdl.K_0:            imgui.Key0,
		sdl.K_1:            imgui.Key1,
		sdl.K_2:            imgui.Key2,
		sdl.K_3:            imgui.Key3,
		sdl.K_4:            imgui.Key4,
		sdl.K_5:            imgui.Key5,
		sdl.K_6:            imgui.Key6,
		sdl.K_7:            imgui.Key7,
		sdl.K_8:            imgui.Key8,
		sdl.K_9:            imgui.Key9,
		sdl.K_a:            imgui.KeyA,
		sdl.K_b:            imgui.KeyB,
		sdl.K_c:            imgui.KeyC,
		sdl.K_d:            imgui.KeyD,
		sdl.K_e:            imgui.KeyE,
		sdl.K_f:            imgui.KeyF,
		sdl.K_g:            imgui.KeyG,
		sdl.K_h:            imgui.KeyH,
		sdl.K_i:            imgui.KeyI,
		sdl.K_j:            imgui.KeyJ,
		sdl.K_k:            imgui.KeyK,
		sdl.K_l:            imgui.KeyL,
		sdl.K_m:            imgui.KeyM,
		sdl.K_n:            imgui.KeyN,
		sdl.K_o:            imgui.KeyO,
		sdl.K_p:            imgui.KeyP,
		sdl.K_q:            imgui.KeyQ,
		sdl.K_r:            imgui.KeyR,
		sdl.K_s:            imgui.KeyS,
		sdl.K_t:            imgui.KeyT,
		sdl.K_u:            imgui.KeyU,
		sdl.K_v:            imgui.KeyV,
		sdl.K_w:            imgui.KeyW,
		sdl.K_x:            imgui.KeyX,
		sdl.K_y:            imgui.KeyY,
		sdl.K_z:            imgui.KeyZ,
		sdl.K_F1:           imgui.KeyF1,
		sdl.K_F2:           imgui.KeyF2,
		sdl.K_F3:           imgui.KeyF3,
		sdl.K_F4:           imgui.KeyF4,
		sdl.K_F5:           imgui.KeyF5,
		sdl.K_F6:           imgui.KeyF6,
		sdl.K_F7:           imgui.KeyF7,
		sdl.K_F8:           imgui.KeyF8,
		sdl.K_F9:           imgui.KeyF9,
		sdl.K_F10:          imgui.KeyF10,
		sdl.K_F11:          imgui.KeyF11,
		sdl.K_F12:          imgui.KeyF12,
	}
)
