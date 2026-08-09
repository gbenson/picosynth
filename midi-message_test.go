package picosynth

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestMIDIMessageTypes(t *testing.T) {
	assert.Equal(t, MIDIMessageType(0x80), NoteOff)
	assert.Equal(t, MIDIMessageType(0x90), NoteOn)
	assert.Equal(t, MIDIMessageType(0xb0), ControlChange)
	assert.Equal(t, MIDIMessageType(0xe0), PitchWheel)
}
