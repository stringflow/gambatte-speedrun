package application

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"time"

	imgui "github.com/AllenDang/cimgui-go"
	"github.com/sqweek/dialog"
	"github.com/veandco/go-sdl2/sdl"
)

type GambatteSource struct {
	Gambatte      *Gambatte
	InputManager  *InputManager
	WindowManager *WindowManager

	SetupText string

	Backbuffer   *Backbuffer
	TextElement  *TextElement
	StateElement *StateElement
}

func NewGambatteSource(renderer *Renderer, gambatte *Gambatte, inputManager *InputManager, windowManager *WindowManager, scaleMode sdl.ScaleMode) (*GambatteSource, error) {
	backbuffer := NewBackbuffer(renderer, GbVideoWidth, GbVideoHeight)
	backbuffer.SetScaleMode(scaleMode)

	stateElement := &StateElement{Slot: 1}

	return &GambatteSource{
		Gambatte:      gambatte,
		InputManager:  inputManager,
		WindowManager: windowManager,

		Backbuffer:   backbuffer,
		TextElement:  &TextElement{},
		StateElement: stateElement,
	}, nil
}

func (src *GambatteSource) Close() {
	src.Backbuffer.Close()
}

func (src *GambatteSource) Draw() {
	src.ProcessInputs()

	imgui.PushStyleVarFloat(imgui.StyleVarWindowRounding, 0)
	imgui.PushStyleVarFloat(imgui.StyleVarWindowBorderSize, 0)
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 0, Y: 0})

	viewport := imgui.MainViewport()
	imgui.SetNextWindowPos(viewport.Pos())
	imgui.SetNextWindowSize(viewport.Size())

	imgui.BeginV("gambatte", nil, imgui.WindowFlagsNoDecoration|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize|imgui.WindowFlagsMenuBar)
	imgui.PopStyleVarV(3)

	if src.Gambatte.RomLoaded && src.Gambatte.BiosLoaded {
		src.DrawBackbuffer()
	} else {
		src.DrawSetup()
	}
	src.DrawMenuBar()

	imgui.End()
}

func (src *GambatteSource) ProcessInputs() {
	if src.Gambatte.ResetStage == NotResetting {
		if src.InputManager.IsPressed[PreviousStateSlot] {
			src.StateElement.Advance(-1)
			src.TextElement.Show(fmt.Sprintf("State %d", src.StateElement.Slot))
		}

		if src.InputManager.IsPressed[NextStateSlot] {
			src.StateElement.Advance(+1)
			src.TextElement.Show(fmt.Sprintf("State %d", src.StateElement.Slot))
		}

		if src.InputManager.IsPressed[SaveState] {
			if src.Gambatte.SaveState(src.StateElement.StatePath()) == nil {
				src.TextElement.Show(fmt.Sprintf("State %d saved", src.StateElement.Slot))
			} else {
				src.TextElement.Show("Failure")
			}
		}

		if src.InputManager.IsPressed[LoadState] {
			if src.Gambatte.LoadState(src.StateElement.StatePath()) == nil {
				src.TextElement.Show(fmt.Sprintf("State %d loaded", src.StateElement.Slot))
			} else {
				src.TextElement.Show("Failure")
			}
		}

		if src.InputManager.IsPressed[Reset] {
			src.Gambatte.Reset()
		}
	}
}

func (src *GambatteSource) DrawBackbuffer() {
	if src.Gambatte.ResetStage == ResetDone {
		src.DrawCrc()
	}

	src.Backbuffer.CopyPixelsRGBANative(src.Gambatte.VideoBuffer[:], 0, 0, GbVideoWidth, GbVideoHeight)
	if src.Gambatte.ResetStage == FadeToBlack || src.Gambatte.ResetStage == Stalling {
		src.ApplyFade()
	}

	src.TextElement.Draw(src.Backbuffer)
	src.StateElement.Draw(src.Backbuffer)
	src.Backbuffer.UpdateVRAM()

	imgui.Image(imgui.TextureID(src.Backbuffer.SdlTexture), imgui.ContentRegionAvail())
}

func (src *GambatteSource) DrawCrc() {
	var checksum strings.Builder
	for i := 0; i < 8; i++ {
		checksumPart := (src.Gambatte.Cartridge.Crc32 >> uint(28-i*4)) & 0xf
		checksum.WriteByte("0123456789ABCDEF"[checksumPart])
	}

	// TODO(stringflow): Is printing only the core revision enough?
	src.TextElement.Show(fmt.Sprintf("Reset r%d %s", GambatteRevision(), checksum.String()))
}

func (src *GambatteSource) ApplyFade() {
	samplesPassed := src.Gambatte.FadeSamplesTotal - src.Gambatte.FadeSamplesLeft
	percentage := float32(samplesPassed) / float32(src.Gambatte.FadeSamplesTotal)
	alpha := byte(255 * percentage)
	src.Backbuffer.FillRect(0, 0, GbVideoWidth, GbVideoHeight, color.RGBA{0, 0, 0, alpha})
}

func (src *GambatteSource) DrawSetup() {
	viewportSize := imgui.MainViewport().Size()

	size := imgui.Vec2{X: 300, Y: 105}
	imgui.SetNextWindowPos(imgui.Vec2{X: (viewportSize.X - size.X) / 2, Y: (viewportSize.Y - size.Y) / 2})

	imgui.BeginChildStrV("gambatte-speedrun setup", size, true, imgui.WindowFlagsNoTitleBar|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize)
	imgui.TextWrapped("gambatte-speedrun requires a boot ROM and a game ROM to run.\n\nTypically the boot ROM is called \"gbc_bios.bin\"")

	if src.SetupText != "" {
		imgui.Separator()
		imgui.TextWrapped(src.SetupText)
	}

	imgui.EndChild()
}

func (src *GambatteSource) DrawMenuBar() {
	if imgui.BeginMenuBar() {
		if imgui.BeginMenu("File") {
			if imgui.MenuItemBoolV("Load ROM", "", false, !src.Gambatte.RomLoaded && src.Gambatte.BiosLoaded) {
				romPath, err := dialog.File().Filter("Game Boy ROM Images", "gbc").Load()
				if err == nil {
					if src.Gambatte.LoadRom(romPath) {
						src.SetupText = "ROM loaded successfully"
						src.StateElement.UpdatePath(romPath)
					} else {
						src.SetupText = "Unable to load ROM"
					}
				}
			}

			if imgui.MenuItemBoolV("Close ROM", "", false, src.Gambatte.RomLoaded) {
				src.SetupText = ""
				src.Gambatte.RomLoaded = false
			}

			if imgui.MenuItemBoolV("Load BIOS", "", false, !src.Gambatte.BiosLoaded) {
				biosPath, err := dialog.File().Filter("BIOS", "bin").Load()
				if err == nil {
					if src.Gambatte.LoadBios(biosPath) {
						src.SetupText = "BIOS loaded successfully"
					} else {
						src.SetupText = "Unable to load BIOS"
					}
				}
			}

			imgui.Separator()

			if imgui.MenuItemBool("Quit") {
				sdl.PushEvent(&sdl.QuitEvent{Type: sdl.QUIT})
			}

			imgui.EndMenu()
		}

		if imgui.BeginMenu("Settings") {
			imgui.InternalPushOverrideID(imgui.InternalImHashStr("Settings"))
			if imgui.MenuItemBool("Input...") {
				imgui.OpenPopupStr("Input Settings")
			}

			if imgui.MenuItemBool("Audio...") {
				imgui.OpenPopupStr("Audio Settings")
			}

			if imgui.MenuItemBool("Video...") {
				imgui.OpenPopupStr("Video Settings")
			}
			imgui.PopID()

			imgui.Separator()

			if imgui.BeginMenu("Windows") {
				for _, guiWindow := range src.WindowManager.OptionalWindows {
					isVisible := guiWindow.GetWindow().Visible()
					if imgui.MenuItemBoolPtr(guiWindow.GetName(), "", &isVisible) {
						guiWindow.GetWindow().SetVisible(isVisible)
					}
				}
				imgui.EndMenu()
			}

			imgui.EndMenu()
		}

		imgui.EndMenuBar()
	}
}

type ElementTimer struct {
	DisappearTime time.Time
}

func (et *ElementTimer) Start(duration time.Duration) {
	et.DisappearTime = time.Now().Add(duration)
}

func (et *ElementTimer) ShouldShow() bool {
	return et.DisappearTime.After(time.Now())
}

type TextElement struct {
	Text  string
	Timer ElementTimer
}

func (te *TextElement) Show(text string) {
	te.Text = text
	te.Timer.Start(4 * time.Second)
}

func (te *TextElement) Draw(target *Backbuffer) {
	if !te.Timer.ShouldShow() {
		return
	}

	target.DrawTextCentered(te.Text, 0, GbVideoHeight-10-Font.SpriteHeight, GbVideoWidth, Font.SpriteHeight)
}

type StateElement struct {
	Slot       int
	State      []byte
	Snapshot   []byte
	ValidState bool
	Timer      ElementTimer

	RomDirectory string
	RomFileName  string
}

func (se *StateElement) UpdatePath(romPath string) {
	se.RomDirectory = filepath.Dir(romPath)
	se.RomFileName = filepath.Base(romPath)
	se.RomFileName = se.RomFileName[:strings.IndexRune(se.RomFileName, '.')]

	se.UpdateState()
}

func (se *StateElement) StatePath() string {
	return filepath.Join(se.RomDirectory, fmt.Sprintf("%s_%d.gqs", se.RomFileName, se.Slot))
}

func (se *StateElement) UpdateState() {
	se.ValidState = false

	gqs, err := os.ReadFile(se.StatePath())
	if err != nil {
		return
	}

	signature := gqs[0]
	version := gqs[1]
	if signature != 0xff || version != 0x02 {
		return
	}

	se.ValidState = true
	se.State = gqs

	snapshotSize := int(gqs[3])<<16 | int(gqs[4])<<8 | int(gqs[5])
	se.Snapshot = gqs[6 : 6+snapshotSize]
}

func (se *StateElement) Advance(amount int) {
	minSlot := 1
	maxSlot := 100

	se.Slot += amount
	if se.Slot < minSlot {
		se.Slot = maxSlot
	} else if se.Slot > maxSlot {
		se.Slot = minSlot
	}

	se.UpdateState()
	se.Show()
}

func (se *StateElement) Show() {
	se.Timer.Start(4 * time.Second)
}

func (se *StateElement) Draw(target *Backbuffer) {
	if !se.Timer.ShouldShow() {
		return
	}

	gridSize := 10
	margin := 10
	offsetX := 11
	offsetY := 6
	width := GbVideoWidth >> 2
	height := GbVideoHeight >> 2

	slot := se.Slot - 1
	x := (slot%gridSize)*offsetX + margin
	y := (slot/gridSize)*offsetY + margin

	if se.ValidState {
		target.CopyPixelsRGBANative(se.Snapshot, x, y, width, height)
	} else {
		target.FillRect(x, y, width, height, color.RGBA{0x20, 0x20, 0x20, 0xe0})
		target.DrawTextCentered("EMPTY", x, y, width, height)
	}
}
