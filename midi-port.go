package picosynth

import (
	"gbenson.net/go/picosynth/internal/hw/machine"
	"gbenson.net/go/picosynth/internal/uart"
)

type MIDIPort struct {
	uart *uart.UART
	msgs <-chan MIDIMessage
}

// Open configures a UART to transmit and receive MIDI via the specified pins.
func (p *MIDIPort) Open(tx, rx machine.Pin) error {
	if p.IsOpen() {
		panic("already opened")
	}

	const (
		BaudRate = 31250
		DataBits = 8
		Parity   = machine.ParityNone
		StopBits = 1
	)

	u, err := uart.Open(tx, rx, BaudRate)
	if err != nil {
		return err
	}

	if err := u.SetFormat(DataBits, StopBits, Parity); err != nil {
		return err
	}

	p.uart = u
	return nil
}

// IsOpen reports whether the port has been opened.
func (p *MIDIPort) IsOpen() bool {
	return p.uart != nil
}

// Name implements [worker].
func (p *MIDIPort) Name() string {
	return "midiport"
}

// Run implements [worker].
func (p *MIDIPort) Run() func() error {
	return p.run
}

func (p *MIDIPort) run() error {
	if p.msgs != nil {
		panic("already started")
	}

	msgs := make(chan MIDIMessage, 256)
	defer close(msgs)
	p.msgs = msgs

	for {
		// Read until we have a status byte.
		status, err := p.readByte()
		if err != nil {
			return err
		} else if status&0x80 == 0 {
			continue // not a status byte
		}

		// Ignore messages we don't care about.
		switch MIDIMessageType(status & 0xf0) {
		case NoteOn:
		case NoteOff:
		case ControlChange:
		case PitchWheel:
		default:
			continue
		}

		// Read the data bytes.
		data1, err := p.readByte()
		if err != nil {
			return err
		}
		data2, err := p.readByte()
		if err != nil {
			return err
		}

		msgs <- midiMessage(status, data1, data2)
	}
}

// readByte reads and returns the next byte from the input or any
// error encountered. If readByte returns an error, no input byte
// was consumed, and the returned byte value is undefined.  Bytes
// that are status bytes of system realtime messages (0xf8-0xff)
// are dropped (System realtime messages are all one byte and can
// appear in the middle of other messages.  We don't handle any;
// dropping them here is hacky but simplifies many things!)
func (p *MIDIPort) readByte() (b byte, err error) {
	for {
		b, err = p.uart.ReadByte()
		if err == nil && b >= 0xf8 {
			continue // system realtime messge
		}
		return
	}
}

// Poll returns the next pending msg, or NoMessage if there are none.
func (p *MIDIPort) Poll() MIDIMessage {
	select {
	case msg := <-p.msgs:
		return msg
	default:
		return NoMessage
	}
}
