package picosynth

import "gbenson.net/go/picosynth/internal/fmt"

type MIDIMessage uint32

// NoMessage is returned by [MIDIReader.Poll] when there are no messages.
const NoMessage MIDIMessage = 0

type MIDIMessageType uint8

const (
	NoteOff MIDIMessageType = (8 + iota) << 4
	NoteOn
	polyAftertouch
	ControlChange
	programChange
	channelAftertouch
	PitchWheel
)

func midiMessage(status, data1, data2 uint8) MIDIMessage {
	packed := (uint32(data2) << 16) | (uint32(data1) << 8) | uint32(status)
	return MIDIMessage(packed)
}

// String implements [fmt.Stringer].
func (m MIDIMessage) String() string {
	return fmt.Hex32Stringer("MIDIMessage", m)
}
