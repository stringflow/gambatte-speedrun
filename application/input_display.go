package application

import (
	"image"
	"image/color"
	"slices"

	imgui "github.com/AllenDang/cimgui-go"
	"github.com/stringflow/gambatte-speedrun/assets"
)

const (
	KeySize          = 34
	KeyPressedBit    = 16
	KeyContractedBit = 32
	KeyElongatedBit  = 64

	ArrowSize       = 14
	ArrowPressedBit = 4

	LastLineWidth = 2
)

type Key struct {
	Input       InputId
	SpriteIndex int
	X           float32
	Y           float32
}

type Arrow struct {
	SpriteIndex int
	OffsetX     int
	OffsetY     int
}

type Dpad struct {
	Key            Key
	Arrow          Arrow
	ShiftX         int
	ShiftY         int
	LastLineX      int
	LastLineY      int
	LastLineWidth  int
	LastLineHeight int
}

type DpadState struct {
	TotalShiftX int
	TotalShiftY int
	LengthBit   int
}

type Palette struct {
	Name   string
	Colors color.Palette
}

func NewPalette(name string, colors color.Palette) Palette {
	end := len(colors) - 1
	bgColor := colors[end]
	r, g, b, a := bgColor.RGBA()

	chromaKeyBgColor := color.RGBA{byte(r), byte(g + 1), byte(b), byte(a)}

	colors = append(colors, bgColor)
	colors[end] = chromaKeyBgColor

	return Palette{
		Name:   name,
		Colors: colors,
	}
}

type InputDisplay struct {
	Window       *Window
	InputManager *InputManager

	Palettes        []Palette
	SelectedPalette string

	Keys []Key
	Dpad []Dpad

	KeySheet   *Spritesheet
	ArrowSheet *Spritesheet
	Backbuffer *Backbuffer
}

func NewInputDisplay(window *Window, inputManager *InputManager, palette string) (*InputDisplay, error) {
	keys := []Key{
		{Input: Select, SpriteIndex: 2, X: 3.5, Y: 6.0},
		{Input: Start, SpriteIndex: 3, X: 4.5, Y: 6.0},
		{Input: B, SpriteIndex: 1, X: 5.5, Y: 4.0},
		{Input: A, SpriteIndex: 0, X: 7.0, Y: 3.0},
		{Input: Reset, SpriteIndex: 4, X: 7.0, Y: 1.0 - (6.0 / KeySize)},
	}

	dpad := []Dpad{
		{Key: Key{Input: Up, SpriteIndex: 5, X: 2.0, Y: 2.0}, Arrow: Arrow{SpriteIndex: 0, OffsetX: 10, OffsetY: 8}, ShiftX: 0, ShiftY: -1, LastLineX: 0, LastLineY: KeySize - 2, LastLineWidth: KeySize, LastLineHeight: 2},
		{Key: Key{Input: Down, SpriteIndex: 6, X: 2.0, Y: 4.0}, Arrow: Arrow{SpriteIndex: 1, OffsetX: 10, OffsetY: 6}, ShiftX: 0, ShiftY: 1, LastLineX: 0, LastLineY: 0, LastLineWidth: KeySize, LastLineHeight: 2},
		{Key: Key{Input: Left, SpriteIndex: 7, X: 1.0, Y: 3.0}, Arrow: Arrow{SpriteIndex: 2, OffsetX: 8, OffsetY: 7}, ShiftX: -1, ShiftY: 0, LastLineX: KeySize - 2, LastLineY: 0, LastLineWidth: 2, LastLineHeight: KeySize},
		{Key: Key{Input: Right, SpriteIndex: 8, X: 3.0, Y: 3.0}, Arrow: Arrow{SpriteIndex: 3, OffsetX: 12, OffsetY: 7}, ShiftX: 1, ShiftY: 0, LastLineX: 0, LastLineY: 0, LastLineWidth: 2, LastLineHeight: KeySize},
	}

	palettes := []Palette{
		NewPalette("Brown", color.Palette{color.RGBA{0, 0, 0, 255}, color.RGBA{228, 150, 133, 255}, color.RGBA{228, 150, 133, 255}, color.RGBA{248, 248, 248, 255}}),
		NewPalette("Pastel Mix", color.Palette{color.RGBA{0, 0, 0, 255}, color.RGBA{228, 144, 163, 255}, color.RGBA{228, 144, 163, 255}, color.RGBA{242, 226, 187, 255}}),
		NewPalette("Blue", color.Palette{color.RGBA{0, 0, 0, 255}, color.RGBA{225, 128, 150, 255}, color.RGBA{113, 182, 208, 255}, color.RGBA{248, 248, 248, 255}}),
		NewPalette("Green", color.Palette{color.RGBA{0, 0, 0, 255}, color.RGBA{96, 186, 46, 255}, color.RGBA{96, 186, 46, 255}, color.RGBA{248, 248, 248, 255}}),
		NewPalette("Red", color.Palette{color.RGBA{0, 0, 0, 255}, color.RGBA{131, 198, 86, 255}, color.RGBA{225, 128, 150, 255}, color.RGBA{248, 248, 248, 255}}),
		NewPalette("Orange", color.Palette{color.RGBA{0, 0, 0, 255}, color.RGBA{232, 186, 77, 255}, color.RGBA{232, 186, 77, 255}, color.RGBA{248, 248, 248, 255}}),
		NewPalette("Dark Blue", color.Palette{color.RGBA{0, 0, 0, 255}, color.RGBA{225, 128, 150, 255}, color.RGBA{141, 156, 191, 255}, color.RGBA{248, 248, 248, 255}}),
		NewPalette("Dark Green", color.Palette{color.RGBA{0, 0, 0, 255}, color.RGBA{225, 128, 150, 255}, color.RGBA{131, 198, 86, 255}, color.RGBA{248, 248, 248, 255}}),
		NewPalette("Dark Brown", color.Palette{color.RGBA{78, 38, 28, 255}, color.RGBA{228, 150, 133, 255}, color.RGBA{189, 146, 144, 255}, color.RGBA{241, 216, 206, 255}}),
		NewPalette("Yellow", color.Palette{color.RGBA{0, 0, 0, 255}, color.RGBA{113, 182, 208, 255}, color.RGBA{232, 186, 77, 255}, color.RGBA{248, 248, 248, 255}}),
		NewPalette("Monochrome", color.Palette{color.RGBA{0, 0, 0, 255}, color.RGBA{160, 160, 160, 255}, color.RGBA{160, 160, 160, 255}, color.RGBA{248, 248, 248, 255}}),
		NewPalette("Inverted", color.Palette{color.RGBA{248, 248, 248, 255}, color.RGBA{24, 128, 104, 255}, color.RGBA{24, 128, 104, 255}, color.RGBA{0, 0, 0, 255}}),
	}

	keySheet := NewSpritesheet(Palettize(ParsePNG(assets.KeysPng), palettes[0].Colors), KeySize, KeySize)
	arrowSheet := NewSpritesheet(Palettize(ParsePNG(assets.ArrowsPng), palettes[0].Colors), ArrowSize, ArrowSize)

	windowWidth, windowHeight := window.SdlWindow.GetSize()
	backbuffer := NewBackbuffer(window.Renderer, int(windowWidth), int(windowHeight))

	inputDisplay := &InputDisplay{
		Window:       window,
		InputManager: inputManager,

		Palettes: palettes,
		Keys:     keys,
		Dpad:     dpad,

		KeySheet:   keySheet,
		ArrowSheet: arrowSheet,
		Backbuffer: backbuffer,
	}

	inputDisplay.SetPalette(palette)
	return inputDisplay, nil
}

func (inputDisplay *InputDisplay) SetPalette(name string) {
	index := slices.IndexFunc(inputDisplay.Palettes, func(p Palette) bool { return p.Name == name })
	if index == -1 {
		index = 0
	}

	palette := inputDisplay.Palettes[index]

	inputDisplay.KeySheet.Image.(*image.Paletted).Palette = palette.Colors
	inputDisplay.ArrowSheet.Image.(*image.Paletted).Palette = palette.Colors
	inputDisplay.SelectedPalette = palette.Name
}

func (inputDisplay *InputDisplay) Draw() {
	inputDisplay.UpdateBackbuffer()
	inputDisplay.Backbuffer.UpdateVRAM()

	imgui.PushStyleVarFloat(imgui.StyleVarWindowRounding, 0)
	imgui.PushStyleVarFloat(imgui.StyleVarWindowBorderSize, 0)
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 0, Y: 0})

	viewport := imgui.MainViewport()
	imgui.SetNextWindowPos(viewport.Pos())
	imgui.SetNextWindowSize(viewport.Size())

	imgui.BeginV("Input Display", nil, imgui.WindowFlagsNoDecoration|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize)
	imgui.PopStyleVarV(3)

	imgui.Image(imgui.TextureID(inputDisplay.Backbuffer.SdlTexture), imgui.ContentRegionAvail())

	displayChromaKeyInformation := false

	if imgui.BeginPopupContextWindow() {
		for _, palette := range inputDisplay.Palettes {
			selected := palette.Name == inputDisplay.SelectedPalette

			if imgui.MenuItemBoolPtr(palette.Name, "", &selected) {
				inputDisplay.SetPalette(palette.Name)
			}
		}

		imgui.Separator()
		displayChromaKeyInformation = imgui.MenuItemBool("Chroma Key Infomation")

		imgui.EndPopup()
	}

	if displayChromaKeyInformation {
		imgui.OpenPopupStr("Chroma Key Information")
	}

	unused := true
	imgui.SetNextWindowSize(imgui.Vec2{X: imgui.WindowSize().X - 20, Y: 130})
	if imgui.BeginPopupModalV("Chroma Key Information", &unused, imgui.WindowFlagsNoResize) {
		imgui.TextWrapped("The following settings need to be applied to the chroma key filter for removing the background:\n\nColor: #000100\nSimilarity: 1\nSmoothness: 1")
		imgui.EndPopup()
	}

	imgui.End()
}

func (inputDisplay *InputDisplay) UpdateBackbuffer() {
	inputDisplay.Backbuffer.Clear(inputDisplay.KeySheet.Image.(*image.Paletted).Palette[3])

	inputDisplay.Backbuffer.FillRect(CalcCoord(2.0), CalcCoord(3.0), KeySize, KeySize, inputDisplay.KeySheet.Image.(*image.Paletted).Palette[4])

	for _, key := range inputDisplay.Keys {
		inputDisplay.DrawKey(key)
	}

	dpadState := inputDisplay.GetDpadState()

	for _, dpad := range inputDisplay.Dpad {
		inputDisplay.DrawDpad(dpad, dpadState)
	}
}

func (inputDisplay *InputDisplay) GetDpadState() DpadState {
	totalShiftX := 0
	totalShiftY := 0

	for _, dpad := range inputDisplay.Dpad {
		if inputDisplay.InputManager.IsDown[dpad.Key.Input] {
			totalShiftX += dpad.ShiftX
			totalShiftY += dpad.ShiftY
		}
	}

	lengthBit := 0
	if totalShiftY < 0 {
		lengthBit = KeyElongatedBit
	} else if totalShiftY > 0 {
		lengthBit = KeyContractedBit
	}

	return DpadState{
		TotalShiftX: totalShiftX,
		TotalShiftY: totalShiftY,
		LengthBit:   lengthBit,
	}
}

func (inputDisplay *InputDisplay) DrawKey(key Key) {
	spriteIndex := key.SpriteIndex

	if inputDisplay.InputManager.IsDown[key.Input] {
		spriteIndex |= KeyPressedBit
	}

	keyX := CalcCoord(key.X)
	keyY := CalcCoord(key.Y)

	inputDisplay.Backbuffer.CopyEntireSprite(inputDisplay.KeySheet, spriteIndex, keyX, keyY)
}

func (inputDisplay *InputDisplay) DrawDpad(dpad Dpad, state DpadState) {
	keySpriteIndex := dpad.Key.SpriteIndex
	arrowSpriteIndex := dpad.Arrow.SpriteIndex

	if inputDisplay.InputManager.IsDown[dpad.Key.Input] {
		keySpriteIndex |= KeyPressedBit
		arrowSpriteIndex |= ArrowPressedBit
	} else {
		keySpriteIndex |= state.LengthBit
	}

	keyX := CalcCoord(dpad.Key.X)
	keyY := CalcCoord(dpad.Key.Y)
	inputDisplay.Backbuffer.CopyEntireSprite(inputDisplay.KeySheet, keySpriteIndex, keyX, keyY)

	arrowX := keyX + dpad.Arrow.OffsetX + state.TotalShiftX
	arrowY := keyY + dpad.Arrow.OffsetY + state.TotalShiftY
	inputDisplay.Backbuffer.CopyEntireSprite(inputDisplay.ArrowSheet, arrowSpriteIndex, arrowX, arrowY)

	// NOTE(stringflow): Copies the last line of the key sprite into the dpad middle to make the keys look connected.
	dpadMiddleX := keyX + dpad.LastLineX - (dpad.ShiftX * dpad.LastLineWidth)
	dpadMiddleY := keyY + dpad.LastLineY - (dpad.ShiftY * dpad.LastLineHeight)
	inputDisplay.Backbuffer.CopySpriteSection(inputDisplay.KeySheet, keySpriteIndex, dpad.LastLineX, dpad.LastLineY, dpadMiddleX, dpadMiddleY, dpad.LastLineWidth, dpad.LastLineHeight)
}

func CalcCoord(value float32) int {
	return int(value * float32(KeySize))
}

func (inputDisplay InputDisplay) Close() {
	inputDisplay.Backbuffer.Close()
}

func (inputDisplay InputDisplay) GetWindow() *Window {
	return inputDisplay.Window
}

func (inputDisplay InputDisplay) GetName() string {
	return "Input Display"
}
