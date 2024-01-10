package application

// #cgo LDFLAGS: ${SRCDIR}/../lib/libgambatte.a -static
// #include "gambatte.h"
import "C"
import (
	"hash/crc32"
	"math/rand"
	"os"
	"unsafe"
)

const (
	GbSampleSize          = 4
	GbSamplesPerFrame     = 35112
	GbAudioBufferOverhead = 2064
	GbVideoWidth          = 160
	GbVideoHeight         = 144

	SamplesToFadeFor  = 1234567
	SamplesToStallFor = 101 * (2 << 14)

	LoadResultOk = 0
	GbcMode      = 1
	GbaFlag      = 2
	SgbMode      = 8
	ReadOnlySav  = 16
	NoBios       = 32

	ExpectedBiosSize = 2304
	ExpectedBiosCrc  = 828843416
)

const ()

type ResetStage int

const (
	NotResetting ResetStage = iota
	FadeToBlack
	Stalling
	ResetDone
)

type Cartridge struct {
	Title [16]byte
	Crc32 uint32
}

type Gambatte struct {
	Gb            *C.GB
	VideoBuffer   [GbVideoWidth * GbVideoHeight * 4]byte
	AudioBuffer   [(GbSamplesPerFrame + GbAudioBufferOverhead) * 2 * 2]byte
	CurrentInputs int
	FrameOverflow int

	ResetStage       ResetStage
	FadeSamplesTotal int
	FadeSamplesLeft  int
	StallSamplesLeft int

	RomLoaded  bool
	BiosLoaded bool
	Cartridge  Cartridge
}

func GambatteRevision() int {
	return int(C.gambatte_revision())
}

//export input_callback
func input_callback(data unsafe.Pointer) C.int {
	gambatte := (*Gambatte)(data)
	return C.int(gambatte.CurrentInputs)
}

func NewGambatte() *Gambatte {
	gambatte := &Gambatte{
		Gb: C.gambatte_create(),
	}

	C.gambatte_setinputgetter(gambatte.Gb, (*C.InputGetter)(C.input_callback), unsafe.Pointer(gambatte))

	return gambatte
}

func (gambatte *Gambatte) Close() {
	C.gambatte_destroy(gambatte.Gb)
}

func (gambatte *Gambatte) IsReady() bool {
	return gambatte.RomLoaded && gambatte.BiosLoaded
}

func (gambatte *Gambatte) LoadRom(path string) bool {
	success := int(C.gambatte_load(gambatte.Gb, C.CString(path), GbcMode|GbaFlag)) == LoadResultOk
	gambatte.RomLoaded = success

	if success {
		// TODO(stringflow): check this for errors... but it shouldn't fail
		romData, _ := os.ReadFile(path)

		gambatte.Cartridge = Cartridge{
			Title: ([16]byte)(romData[0x134 : 0x134+16]),
			Crc32: crc32.ChecksumIEEE(romData),
		}
	}

	return success
}

func (gambatte *Gambatte) LoadBios(path string) bool {
	success := C.gambatte_loadbios(gambatte.Gb, C.CString(path), ExpectedBiosSize, ExpectedBiosCrc) == LoadResultOk
	gambatte.BiosLoaded = success
	return success
}

func (gambatte *Gambatte) AdvanceFrame() int {
	samples := GbSamplesPerFrame - gambatte.FrameOverflow
	offset := C.gambatte_runfor(gambatte.Gb, (*C.uint)(unsafe.Pointer(&gambatte.VideoBuffer[0])), GbVideoWidth, (*C.uint)(unsafe.Pointer(&gambatte.AudioBuffer[0])), (*C.uint)(unsafe.Pointer(&samples)))
	samplesElapsed := int(samples)

	gambatte.FrameOverflow += samplesElapsed

	if offset >= 0 || gambatte.FrameOverflow >= GbSamplesPerFrame {
		gambatte.FrameOverflow = 0
	}

	gambatte.HandleReset(samplesElapsed)

	return samplesElapsed
}

func (gambatte *Gambatte) HandleReset(samplesElapsed int) {
	if gambatte.ResetStage == ResetDone {
		gambatte.ResetStage = NotResetting
	} else if gambatte.ResetStage == FadeToBlack {
		gambatte.HandleFade(samplesElapsed)
	} else if gambatte.ResetStage == Stalling {
		gambatte.HandleStalling(samplesElapsed)
	}
}

func (gambatte *Gambatte) HandleFade(samplesElapsed int) {
	gambatte.FadeSamplesLeft -= samplesElapsed

	if gambatte.FadeSamplesLeft <= 0 {
		gambatte.FadeSamplesLeft = 0

		C.gambatte_reset(gambatte.Gb, (C.uint)(SamplesToStallFor))
		gambatte.StallSamplesLeft = SamplesToStallFor
		gambatte.ResetStage = Stalling
	}
}

func (gambatte *Gambatte) HandleStalling(samplesElapsed int) {
	gambatte.StallSamplesLeft -= samplesElapsed

	if gambatte.StallSamplesLeft <= 0 {
		gambatte.StallSamplesLeft = 0
		gambatte.ResetStage = ResetDone
	}
}

func (gambatte *Gambatte) SaveState(path string) error {
	size := int(C.gambatte_savestate(gambatte.Gb, nil, GbVideoWidth, nil))

	stateBuffer := make([]byte, size)
	C.gambatte_savestate(gambatte.Gb, (*C.uint)(unsafe.Pointer(&gambatte.VideoBuffer[0])), GbVideoWidth, (*C.char)(unsafe.Pointer(&stateBuffer[0])))

	return os.WriteFile(path, stateBuffer, 0644)
}

func (gambatte *Gambatte) LoadState(path string) error {
	stateBuffer, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	success := C.gambatte_loadstate(gambatte.Gb, (*C.char)(unsafe.Pointer(&stateBuffer[0])), (C.uint)(len(stateBuffer)))
	if success == 0 {
		return os.ErrInvalid
	}

	return nil
}

func (gambatte *Gambatte) Reset() {
	samples := SamplesToFadeFor + rand.Intn(GbSamplesPerFrame)

	gambatte.ResetStage = FadeToBlack
	gambatte.FadeSamplesTotal = samples
	gambatte.FadeSamplesLeft = samples
}
