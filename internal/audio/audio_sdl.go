//go:build !tinygo

package audio

// typedef unsigned char Uint8;
// void fillBuffer(void *userdata, Uint8 *stream, int len);
import "C"
import (
	"fmt"
	"runtime"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

type device struct {
	sampleRate   int
	started      bool
	outC         chan []uint16
	errC         chan error
	pinner       runtime.Pinner
	deviceID     sdl.AudioDeviceID
	deviceOpened bool
}

func open(sampleRate int) (Device, error) {
	return &device{sampleRate: sampleRate}, nil
}

// WriteMono implements [Device].
func (d *device) WriteMono(buf []uint16) error {
	if !d.started {
		if err := d.start(len(buf)); err != nil {
			return err
		}
	}

	d.outC <- buf
	return <-d.errC
}

func (sink *device) start(bufferFrames int) error {
	if sink.started {
		panic("already started")
	}
	sink.started = true

	if sdl.WasInit(sdl.INIT_AUDIO) == 0 {
		if err := sdl.InitSubSystem(sdl.INIT_AUDIO); err != nil {
			return err
		}
	}

	sink.outC = make(chan []uint16)
	sink.errC = make(chan error)
	sink.pinner.Pin(sink)

	desiredSpec := sdl.AudioSpec{
		Freq:     int32(sink.sampleRate),
		Format:   sdl.AUDIO_U16SYS,
		Channels: 1,
		Samples:  uint16(bufferFrames),
		Callback: sdl.AudioCallback(C.fillBuffer),
		UserData: unsafe.Pointer(sink),
	}

	var spec sdl.AudioSpec
	dev, err := sdl.OpenAudioDevice("", false, &desiredSpec, &spec, 0)
	if err != nil {
		return err
	}
	sink.deviceID = dev
	sink.deviceOpened = true

	if uint(spec.Freq) != uint(sink.sampleRate) {
		return fmt.Errorf("unexpected sample rate (%d)", spec.Freq)
	} else if spec.Format != desiredSpec.Format {
		return fmt.Errorf("unexpected sample format 0x%x", spec.Format)
	} else if spec.Channels != desiredSpec.Channels {
		return fmt.Errorf("unexpected number of channels (%d)", spec.Channels)
	} else if int(spec.Samples) != bufferFrames {
		return fmt.Errorf("unexpected buffer size %d", spec.Samples)
	}

	sdl.PauseAudioDevice(sink.deviceID, false)
	return nil
}

// Close implements [io.Closer].
func (sink *device) Close() error {
	if !sink.started {
		return nil
	}

	if sink.outC == nil {
		return nil
	}
	defer func() { sink.outC = nil }()
	defer close(sink.outC)
	defer close(sink.errC)

	defer sink.pinner.Unpin()

	if !sink.deviceOpened {
		return nil
	}
	defer func() { sink.deviceOpened = false }()
	defer sdl.CloseAudioDevice(sink.deviceID)

	return nil
}

//export fillBuffer
func fillBuffer(sinkPtr unsafe.Pointer, stream *C.Uint8, length C.int) {
	sink := (*device)(sinkPtr)

	src := <-sink.outC
	dst := unsafe.Slice((*uint16)(unsafe.Pointer(stream)), length/2)
	copy(dst, src)

	sink.errC <- nil
}
