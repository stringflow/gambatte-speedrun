package application

import (
	"fmt"
	"time"

	imgui "github.com/AllenDang/cimgui-go"
	"github.com/veandco/go-sdl2/sdl"
)

type SettingsModal struct {
	InputManager   *InputManager
	WindowManager  *WindowManager
	Audio          *Audio
	GambatteSource *GambatteSource
	GambatteDriver *GambatteDriver

	TargetInput *Input
}

func NewSettingsModal(inputManager *InputManager, audio *Audio, windowManager *WindowManager, gambatteSource *GambatteSource, gambatteDriver *GambatteDriver) *SettingsModal {
	return &SettingsModal{
		InputManager:   inputManager,
		Audio:          audio,
		WindowManager:  windowManager,
		GambatteSource: gambatteSource,
		GambatteDriver: gambatteDriver,
	}
}

func (sm *SettingsModal) Draw() {
	imgui.InternalPushOverrideID(imgui.InternalImHashStr("Settings"))

	sm.DrawInputSettings()
	sm.DrawAudioSettings()
	sm.DrawVideoSettings()

	imgui.PopID()
}

func (sm *SettingsModal) DrawInputSettings() {
	unused := true

	imgui.SetNextWindowSize(imgui.Vec2{X: 265, Y: 0})
	if imgui.BeginPopupModalV("Input Settings", &unused, imgui.WindowFlagsNoResize) {
		if imgui.FrameCount()%60 == 0 {
			sm.InputManager.QueryJoysticks()
		}

		if imgui.BeginTabBar("Input Settings Tabs") {
			if imgui.BeginTabItem("Game") {
				sm.DrawInputTab(A, []string{"A", "B", "Select", "Start", "Right", "Left", "Up", "Down", "Hard Reset"})
				imgui.EndTabItem()
			}

			if imgui.BeginTabItem("State") {
				sm.DrawInputTab(SaveState, []string{"Save State", "Load State", "Prev State", "Next State"})
				imgui.EndTabItem()
			}

			imgui.EndTabBar()
		}

		imgui.EndPopup()
	}
}

func (sm *SettingsModal) DrawInputTab(start InputId, names []string) {
	sm.DrawController()
	sm.DrawInputButtons(names, start)
	sm.DrawInputModal()
}

func (sm *SettingsModal) DrawController() {
	currentJoystick := sm.InputManager.OpenedJoystickName

	if imgui.BeginCombo("Controller", currentJoystick) {
		for _, joystickName := range sm.InputManager.AvailableJoysticks {
			if imgui.SelectableBoolV(joystickName, currentJoystick == joystickName, 0, imgui.Vec2{}) {
				sm.InputManager.OpenJoystick(joystickName)
			}
		}

		imgui.EndCombo()
	}
}

func (sm *SettingsModal) DrawInputButtons(names []string, startId InputId) {
	for i := 0; i < len(names); i++ {
		inputId := InputId(i) + startId

		if imgui.ButtonV(sm.InputManager.Inputs[inputId].GetName(), imgui.Vec2{X: 170, Y: 0}) {
			sm.TargetInput = &sm.InputManager.Inputs[inputId]
			imgui.OpenPopupStr("Input Mapping")
		}

		imgui.SameLine()
		imgui.Text(names[i])
	}
}

func (sm *SettingsModal) DrawInputModal() {
	if imgui.BeginPopup("Input Mapping") && sm.TargetInput != nil {
		for keysym, down := range sm.InputManager.Keyboard {
			if down {
				*sm.TargetInput = NewKeyInput(keysym, sdl.KMOD_NONE)
				imgui.CloseCurrentPopup()
			}
		}

		for button, down := range sm.InputManager.JoyButtons {
			if down {
				*sm.TargetInput = NewJoyButtonInput(button)
				imgui.CloseCurrentPopup()
			}
		}

		for axis, down := range sm.InputManager.JoyAxes {
			if down.Positive || down.Negative {
				*sm.TargetInput = NewJoyAxisInput(axis, down.Positive)
				imgui.CloseCurrentPopup()
			}
		}

		imgui.Text("Press a key or button to map to this input")
		imgui.EndPopup()
	} else {
		sm.TargetInput = nil
	}
}

func (sm *SettingsModal) DrawVideoSettings() {
	unused := true

	imgui.SetNextWindowSize(imgui.Vec2{X: 265, Y: 0})
	if imgui.BeginPopupModalV("Video Settings", &unused, imgui.WindowFlagsNoResize) {
		imgui.PushItemWidth(160)
		sm.DrawGameScale()
		sm.DrawScaleMode()
		imgui.PopItemWidth()
		imgui.EndPopup()
	}
}

func (sm *SettingsModal) DrawGameScale() {
	currentScale := sm.WindowManager.MainWindow.Scale

	if imgui.BeginCombo("Game Scale", ScaleToResolutionString(currentScale)) {
		for scale := int32(1); scale <= 8; scale++ {
			if imgui.SelectableBoolV(ScaleToResolutionString(scale), scale == currentScale, 0, imgui.Vec2{}) {
				sm.WindowManager.MainWindow.SetScale(scale)
			}
		}

		imgui.EndCombo()
	}
}

func ScaleToResolutionString(scale int32) string {
	return fmt.Sprintf("%dx%d", scale*GbVideoWidth, scale*GbVideoHeight)
}

func (sm *SettingsModal) DrawScaleMode() {
	scaleModes := []sdl.ScaleMode{sdl.ScaleModeNearest, sdl.ScaleModeLinear, sdl.ScaleModeBest}
	currentScaleMode := sm.GambatteSource.Backbuffer.ScaleMode()

	if imgui.BeginCombo("Scale Filter", NameForScaleMode(currentScaleMode)) {
		for _, scaleMode := range scaleModes {
			if imgui.SelectableBoolV(NameForScaleMode(scaleMode), scaleMode == currentScaleMode, 0, imgui.Vec2{}) {
				sm.GambatteSource.Backbuffer.SetScaleMode(scaleMode)
			}
		}

		imgui.EndCombo()
	}
}

func NameForScaleMode(scaleMode sdl.ScaleMode) string {
	switch scaleMode {
	case sdl.ScaleModeNearest:
		return "Nearest"
	case sdl.ScaleModeLinear:
		return "Linear"
	case sdl.ScaleModeBest:
		return "Best"
	default:
		return "Unknown"
	}
}

func (sm *SettingsModal) DrawAudioSettings() {
	unused := true

	imgui.SetNextWindowSize(imgui.Vec2{X: 300, Y: 0})
	if imgui.BeginPopupModalV("Audio Settings", &unused, imgui.WindowFlagsNoResize) {
		if imgui.FrameCount()%60 == 0 {
			sm.Audio.QueryDevices()
		}

		sm.DrawAudioDevice()
		imgui.PushItemWidth(100)
		sm.DrawBufferSize()
		sm.DrawVolume()
		imgui.PopItemWidth()
		imgui.EndPopup()
	}
}

func (sm *SettingsModal) DrawAudioDevice() {
	currentDevice := sm.Audio.OpenedDeviceName

	if imgui.BeginCombo("Audio Device", currentDevice) {
		devices := []string{SystemDefaultAudioDevice}
		devices = append(devices, sm.Audio.AvailableDevices...)

		for _, deviceName := range devices {
			if imgui.SelectableBoolV(deviceName, currentDevice == deviceName, 0, imgui.Vec2{}) {
				sm.Audio.OpenDevice(deviceName)
			}
		}

		imgui.EndCombo()
	}
}

func (sm *SettingsModal) DrawBufferSize() {
	bufferSize := int32(sm.GambatteDriver.BufferSize / time.Millisecond)

	if imgui.InputInt("Buffer Size (ms)", &bufferSize) {
		bufferSize = Clamp(bufferSize, 10, 100)
		sm.GambatteDriver.BufferSize = time.Duration(bufferSize) * time.Millisecond
	}
}

func (sm *SettingsModal) DrawVolume() {
	if imgui.InputInt("Volume", &sm.Audio.Volume) {
		sm.Audio.Volume = Clamp(sm.Audio.Volume, 0, 100)
	}
}
