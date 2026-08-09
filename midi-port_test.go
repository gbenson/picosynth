package picosynth

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"gbenson.net/go/picosynth/internal/hw/machine"
)

func TestMIDIPort(t *testing.T) {
	WithTestEmulator(t, &midiTestEmulator{})

	p := &MIDIPort{}
	assert.Check(t, !p.IsOpen())
	assert.NilError(t, p.Open(machine.GP1, machine.GP14))
	assert.Check(t, p.IsOpen())
	assert.Equal(t, p.Poll(), NoMessage)

	ctx := t.Context()
	var msgs []MIDIMessage
	go func() {
		for {
			if ctx.Err() != nil {
				break
			}
			msg := p.Poll()
			if msg == NoMessage {
				time.Sleep(time.Millisecond)
				continue
			}
			msgs = append(msgs, msg)
		}
	}()

	assert.Equal(t, p.run(), ErrTestComplete)
	time.Sleep(time.Millisecond)
	assert.Equal(t, p.Poll(), NoMessage)
	assert.Equal(t, len(msgs), 4)

	assert.Equal(t, msgs[0], MIDIMessage(0x463c90)) // note on
	assert.Equal(t, msgs[1], MIDIMessage(0x243c80)) // note off
	assert.Equal(t, msgs[2], MIDIMessage(0x4c01b0)) // mod wheel
	assert.Equal(t, msgs[3], MIDIMessage(0x1319e0)) // pitch bend
}

type midiTestEmulator struct {
	pos int
}

func (e *midiTestEmulator) UARTBuffered(uart machine.UART) int {
	return 123
}

func (e *midiTestEmulator) UARTReadByte(uart machine.UART) (byte, error) {
	p := e.pos
	if p >= len(midiTestData) {
		return 0, ErrTestComplete
	}
	e.pos = p + 1
	return midiTestData[p], nil
}

var midiTestData = []byte{
	0x90, 0x3c, 0x46, // note on
	0x80, 0x3c, 0x24, // note off
	0xb0, 0x01, 0x4c, // mod wheel
	0xe0, 0x19, 0x13, // pitch bend
}
