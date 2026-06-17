package audio

import "gbenson.net/go/picosynth/internal/hw"

type Device = hw.AudioDevice

// Open opens an output device with the specified sample rate.
func Open(sampleRate int) (Device, error) {
	return open(sampleRate)
}
