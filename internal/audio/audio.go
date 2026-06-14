package audio

import "io"

type Device interface {
	io.Closer

	// WriteMono writes a mono audio buffer to the audio device.
	WriteMono(buf []int16) error
}

// Open opens an output device with the specified sample rate.
func Open(sampleRate int) (Device, error) {
	return open(sampleRate)
}
