package application

import (
	"time"
)

type GambatteDriver struct {
	Gambatte     *Gambatte
	Audio        *Audio
	InputManager *InputManager
	BufferSize   time.Duration
}

func NewGambatteDriver(gambatte *Gambatte, audio *Audio, inputManager *InputManager, bufferSize time.Duration) *GambatteDriver {
	return &GambatteDriver{
		Gambatte:     gambatte,
		Audio:        audio,
		InputManager: inputManager,
		BufferSize:   bufferSize,
	}
}

func (driver GambatteDriver) UpdateEmulator() {
	if driver.Audio.GetQueuedDuration() < driver.BufferSize {
		driver.UpdateJoypad()
		samplesToPlay := driver.Gambatte.AdvanceFrame()
		driver.PlayAudio(samplesToPlay)
	} else {
		time.Sleep(driver.BufferSize / 4)
	}
}

func (driver GambatteDriver) UpdateJoypad() {
	driver.Gambatte.CurrentInputs = 0

	for input := A; input <= Down; input++ {
		if driver.InputManager.IsDown[input] {
			driver.Gambatte.CurrentInputs |= 1 << uint(input)
		}
	}
}

func (driver GambatteDriver) PlayAudio(samplesToPlay int) {
	bytesToPlay := samplesToPlay * GbSampleSize
	driver.Audio.QueueSamples(driver.Gambatte.AudioBuffer[:bytesToPlay])
}
