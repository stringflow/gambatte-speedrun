package application

import (
	"slices"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

const SystemDefaultAudioDevice = "[System Default]"
const GbSamplesPerSecond = 2097152

type Audio struct {
	Device    sdl.AudioDeviceID
	Spec      sdl.AudioSpec
	Resampler *sdl.AudioStream

	AvailableDevices []string
	OpenedDeviceName string
	Volume           int32
}

func NewAudio(deviceName string, volume int32) (*Audio, error) {
	audio := &Audio{
		Volume: volume,
	}
	audio.QueryDevices()
	err := audio.OpenDevice(deviceName)
	return audio, err
}

func (audio *Audio) OpenDevice(desiredDeviceName string) error {
	if audio.Resampler != nil {
		audio.CloseCurrentDevice()
	}

	obtainedDeviceName := ""
	if desiredDeviceName != SystemDefaultAudioDevice && slices.Contains(audio.AvailableDevices, desiredDeviceName) {
		obtainedDeviceName = desiredDeviceName
	}

	desiredSpec := sdl.AudioSpec{
		Freq:     48000,
		Format:   sdl.AUDIO_S16LSB,
		Channels: 2,
		Samples:  32,
		Callback: nil,
	}

	obtainedSpec := sdl.AudioSpec{}
	device, err := sdl.OpenAudioDevice(obtainedDeviceName, false, &desiredSpec, &obtainedSpec, sdl.AUDIO_ALLOW_ANY_CHANGE)
	if err != nil {
		return err
	}

	resampler, err := sdl.NewAudioStream(sdl.AUDIO_S16LSB, 2, GbSamplesPerSecond, obtainedSpec.Format, obtainedSpec.Channels, int(obtainedSpec.Freq))
	if err != nil {
		return err
	}

	openedDeviceName := obtainedDeviceName
	if openedDeviceName == "" {
		openedDeviceName = SystemDefaultAudioDevice
	}

	audio.Device = device
	audio.Resampler = resampler
	audio.Spec = obtainedSpec
	audio.OpenedDeviceName = openedDeviceName
	sdl.PauseAudioDevice(device, false)

	return nil
}

func (audio *Audio) Close() {
	audio.CloseCurrentDevice()
}

func (audio *Audio) CloseCurrentDevice() {
	audio.Resampler.Free()
	sdl.CloseAudioDevice(audio.Device)
}

func (audio *Audio) QueryDevices() {
	audio.AvailableDevices = make([]string, sdl.GetNumAudioDevices(false))
	for i := range audio.AvailableDevices {
		audio.AvailableDevices[i] = sdl.GetAudioDeviceName(i, false)
	}
}

func (audio *Audio) QueueSamples(samples []byte) error {
	err := audio.Resampler.Put(samples)
	if err != nil {
		return err
	}

	available := audio.Resampler.Available()
	if available > 0 {
		resampled := make([]byte, available)
		_, err := audio.Resampler.Get(resampled)
		if err != nil {
			return err
		}

		volumeAdjustedSamples := audio.ChangeVolumeOfSamples(resampled, audio.Volume)
		sdl.QueueAudio(audio.Device, volumeAdjustedSamples)
	}

	return nil
}

func (audio *Audio) ChangeVolumeOfSamples(samples []byte, volume int32) []byte {
	// NOTE(stringflow): Unfortunately SDL does not let us use the same buffer as both the source and destination.
	// So we have to make a copy.
	size := len(samples)
	volumeAdjustedSamples := make([]byte, size)
	sdlVolume := Clamp(float32(volume)/100.0, 0.0, 1.0) * sdl.MIX_MAXVOLUME
	sdl.MixAudioFormat(&volumeAdjustedSamples[0], &samples[0], audio.Spec.Format, uint32(size), int(sdlVolume))

	return volumeAdjustedSamples
}

func (audio *Audio) GetQueuedDuration() time.Duration {
	queuedSamples := float64(sdl.GetQueuedAudioSize(audio.Device)) / float64(audio.Spec.Format.BitSize()/8*audio.Spec.Channels)
	queuedSeconds := queuedSamples / float64(audio.Spec.Freq)
	return time.Duration(queuedSeconds * float64(time.Second))
}
