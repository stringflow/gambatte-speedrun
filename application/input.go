package application

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"

	"github.com/veandco/go-sdl2/sdl"
)

type InputId int
type InputArray [InputCount]Input

const (
	A InputId = iota
	B
	Select
	Start
	Right
	Left
	Up
	Down
	Reset
	SaveState
	LoadState
	PreviousStateSlot
	NextStateSlot

	InputCount
)

type AxisState struct {
	Positive bool
	Negative bool
}

type InputManager struct {
	Inputs       InputArray
	Builtins     InputArray
	IsDown       [InputCount]bool // NOTE(stringflow): Will stay true if input is held down on consecutive frames
	IsPressed    [InputCount]bool // NOTE(stringflow): Will only be true on the frame the input is first pressed
	DownPrevious [InputCount]bool

	Keyboard     map[sdl.Keycode]bool
	Keymodifiers sdl.Keymod
	JoyButtons   map[byte]bool
	JoyAxes      map[byte]AxisState

	OpenedJoystick     *sdl.Joystick
	OpenedJoystickName string
	AvailableJoysticks []string
}

func NewInputManager(inputs InputArray, joystickName string) *InputManager {
	inputManager := &InputManager{
		Inputs: inputs,
		Builtins: [InputCount]Input{
			Reset:             NewKeyInput(sdl.K_r, sdl.KMOD_CTRL),
			SaveState:         NewKeyInput(sdl.K_s, sdl.KMOD_CTRL),
			LoadState:         NewKeyInput(sdl.K_l, sdl.KMOD_CTRL),
			PreviousStateSlot: NewKeyInput(sdl.K_z, sdl.KMOD_CTRL),
			NextStateSlot:     NewKeyInput(sdl.K_x, sdl.KMOD_CTRL),
		},

		Keyboard:     make(map[sdl.Keycode]bool),
		Keymodifiers: sdl.KMOD_NONE,
		JoyButtons:   make(map[byte]bool),
		JoyAxes:      make(map[byte]AxisState),
	}

	inputManager.QueryJoysticks()
	inputManager.OpenJoystick(joystickName)

	return inputManager
}

func (manager *InputManager) Close() {
	manager.CloseCurrentJoystick()
}

func (manager *InputManager) OpenJoystick(name string) {
	manager.CloseCurrentJoystick()

	joystickId := slices.Index(manager.AvailableJoysticks, name)
	if joystickId == -1 {
		if len(manager.AvailableJoysticks) > 0 {
			joystickId = 0
		} else {
			manager.OpenedJoystick = nil
			manager.OpenedJoystickName = ""
			return
		}
	}

	joystick := sdl.JoystickOpen(joystickId)
	manager.OpenedJoystick = joystick
	manager.OpenedJoystickName = joystick.Name()
}

func (manager *InputManager) CloseCurrentJoystick() {
	if manager.OpenedJoystick != nil {
		manager.OpenedJoystick.Close()
	}
}

func (manager *InputManager) QueryJoysticks() {
	manager.AvailableJoysticks = make([]string, sdl.NumJoysticks())
	for i := range manager.AvailableJoysticks {
		manager.AvailableJoysticks[i] = sdl.JoystickNameForIndex(i)
	}
}

func (manager *InputManager) ProcessEvent(event sdl.Event) {
	switch e := event.(type) {
	case sdl.KeyboardEvent:
		manager.Keyboard[e.Keysym.Sym] = e.State == sdl.PRESSED
	case sdl.JoyButtonEvent:
		manager.JoyButtons[e.Button] = e.State == sdl.PRESSED
	case sdl.JoyAxisEvent:
		initialState, ok := manager.OpenedJoystick.AxisInitialState(int(e.Axis))

		if ok && initialState == math.MinInt16 {
			// NOTE(stringflow): trigger button
			const threshold int16 = 0
			manager.JoyAxes[e.Axis] = AxisState{Positive: e.Value > threshold, Negative: false}
		} else {
			// NOTE(stringflow): analog stick
			const threshold int16 = 8000
			manager.JoyAxes[e.Axis] = AxisState{Positive: e.Value > threshold, Negative: e.Value < -threshold}
		}
	}
}

func (manager *InputManager) Update() {
	manager.Keymodifiers = sdl.GetModState()

	for input := A; input < InputCount; input++ {
		isDown := false
		isDown = isDown || manager.Inputs[input].IsDown(manager)

		builtin := manager.Builtins[input]
		if builtin != nil {
			isDown = isDown || builtin.IsDown(manager)
		}

		manager.IsDown[input] = isDown
		manager.IsPressed[input] = isDown && !manager.DownPrevious[input]
	}

	manager.DownPrevious = manager.IsDown
}

type Input interface {
	IsDown(inputManager *InputManager) bool
	GetName() string
}

func (inputs *InputArray) UnmarshalJSON(data []byte) error {
	var inputList []json.RawMessage
	err := json.Unmarshal(data, &inputList)
	if err != nil {
		return err
	}

	for i, inputJson := range inputList {
		var input map[string]interface{}
		err := json.Unmarshal(inputJson, &input)
		if err != nil {
			return err
		}

		_, ok := input["Keycode"]
		if ok {
			inputs[i] = NewKeyInput(sdl.Keycode(input["Keycode"].(float64)), sdl.Keymod(input["Modifiers"].(float64)))
			continue
		}

		_, ok = input["Button"]
		if ok {
			inputs[i] = NewJoyButtonInput(byte(input["Button"].(float64)))
			continue
		}

		_, ok = input["Axis"]
		if ok {
			inputs[i] = NewJoyAxisInput(byte(input["Axis"].(float64)), input["Positive"].(bool))
			continue
		}
	}

	return nil
}

type BaseInput struct {
	Name string
}

func (input BaseInput) GetName() string {
	return input.Name
}

type KeyInput struct {
	BaseInput
	Keycode   sdl.Keycode
	Modifiers sdl.Keymod
}

func NewKeyInput(keycode sdl.Keycode, modifiers sdl.Keymod) KeyInput {
	return KeyInput{
		BaseInput: BaseInput{
			Name: sdl.GetKeyName(keycode),
		},
		Keycode:   keycode,
		Modifiers: modifiers,
	}
}

func (input KeyInput) IsDown(manager *InputManager) bool {
	isKeysymDown := manager.Keyboard[input.Keycode]
	areModifiersRequired := input.Modifiers != sdl.KMOD_NONE
	areModifierMatching := (manager.Keymodifiers & input.Modifiers) != 0
	return isKeysymDown && (!areModifiersRequired || areModifierMatching)
}

type JoyButtonInput struct {
	BaseInput
	Button byte
}

func NewJoyButtonInput(button byte) JoyButtonInput {
	return JoyButtonInput{
		BaseInput: BaseInput{
			Name: fmt.Sprintf("Button %d", button),
		},
		Button: button,
	}
}

func (input JoyButtonInput) IsDown(manager *InputManager) bool {
	return manager.JoyButtons[input.Button]
}

type JoyAxisInput struct {
	BaseInput
	Axis     byte
	Positive bool
}

func NewJoyAxisInput(axis byte, positive bool) JoyAxisInput {
	positiveString := ""
	if positive {
		positiveString = "+"
	} else {
		positiveString = "-"
	}

	return JoyAxisInput{
		BaseInput: BaseInput{
			Name: fmt.Sprintf("Axis %d %s", axis, positiveString),
		},
		Axis:     axis,
		Positive: positive,
	}
}

func (input JoyAxisInput) IsDown(manager *InputManager) bool {
	stickState := manager.JoyAxes[input.Axis]
	if input.Positive {
		return stickState.Positive
	} else {
		return stickState.Negative
	}
}
