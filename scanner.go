package picosynth

import (
	"math/bits"
	"time"

	"gbenson.net/go/picosynth/internal/adc"
	"gbenson.net/go/picosynth/internal/encoder"
	"gbenson.net/go/picosynth/internal/keyboard"
)

const (
	SampleRate = 48000
	MaxLatency = 10 * time.Millisecond

	// Split MaxLatency between scanning inputs and generating audio.
	// maxLatency is *half* of MaxLatency because:
	//
	//  - The scanner takes the first part, the limiting case being a
	//    key or button changing state an instant after being scanned,
	//    so a full loop has to complete before a corresponding event
	//    is emitted.
	//
	//  - The time taken to empty the audio buffer takes up the second
	//    part, the limiting case being that the first sample generated
	//    into every buffer has to wait for the entire other buffer to
	//    play before being output.
	maxLatency = MaxLatency / 2

	// maxScanBufferFrames is the maximum number of audio frames we
	// can buffer while scanning the entire keyboard matrix without
	// exceeding maxLatency.
	maxScanBufferFrames = int(SampleRate * maxLatency / time.Second)

	// numKeyboardMatrixCols, numKeyboardMatrixRows define the size
	// of the keyboard matrix.
	numKeyboardMatrixCols = 8 // inputs
	numKeyboardMatrixRows = 7 // outputs

	// maxRowBufferFrames is the maximum number of audio frames we can
	// buffer while scanning one row of the keyboard matrix without
	// exceeding maxLatency.
	maxRowBufferFrames = maxScanBufferFrames / numKeyboardMatrixRows
)

var (
	// BufferFrames is the size of buffers passed to [Engine.Fill].
	// The SDL emulator requires this to be a power of two.
	BufferFrames = 1 << (bits.Len(uint(maxRowBufferFrames)) - 1)

	// TickRate is the rate at which [Scanner.Scan] should be called.
	TickRate = SampleRate / BufferFrames
)

// Scanner reads all input from the user.
type Scanner struct {
	rows    [numKeyboardMatrixRows]uint8
	adcs    [4]adc.ADC       // Pico has 4 ADCs.
	encoder *encoder.Encoder // Picosynth has 1 encoder.

	row int // which keyboard matrix row will we scan next?
	adc int // which ADC will we scan next?

	// Each Scan could emit up to:
	//  - 8 key up/down events
	//  - 1 encoder event
	//  - 1 ADC reading
	eventBuf [numKeyboardMatrixCols + 2]Event
}

func (s *Scanner) init() error {
	if len(keyboard.Matrix.Inputs) != numKeyboardMatrixCols {
		panic("mismatch (columns)")
	}
	if len(keyboard.Matrix.Outputs) != numKeyboardMatrixRows {
		panic("mismatch (rows)")
	}

	for i := range s.adcs {
		if a, err := adc.Open(i); err != nil {
			return err
		} else {
			s.adcs[i] = a
		}
	}

	if e, err := encoder.Open(0); err != nil {
		return err
	} else {
		s.encoder = e
	}

	return nil
}

func (s *Scanner) Scan() []Event {
	events := s.eventBuf[:]

	events = s.scanKeys(events)
	events = s.scanADCs(events)
	events = s.scanEncoder(events)

	return events
}

// scan one row of the keyboard matrix per tick.
func (s *Scanner) scanKeys(events []Event) []Event {
	ksm := keyboard.Matrix
	rows := ksm.Outputs
	cols := ksm.Inputs

	rowIndex := s.row
	oldState := s.rows[rowIndex]
	newState := uint8(0)

	// scan the row
	colMask := uint8(1)
	for _, col := range cols {
		if col.Get() {
			newState |= colMask
		}
		colMask <<= 1
	}

	// handle changes
	if delta := oldState ^ newState; delta != 0 {
		rowBase := Scancode(rowIndex * numKeyboardMatrixCols)
		colMask := uint8(1)
		for colIndex := range numKeyboardMatrixCols {
			if delta&colMask != 0 {
				code := rowBase + Scancode(colIndex)
				if newState&colMask != 0 {
					events = append(events, NewKeyDownEvent(code))
				} else {
					events = append(events, NewKeyUpEvent(code))
				}
			}
			colMask <<= 1
		}

		s.rows[rowIndex] = newState
	}

	// de-energize the row we just scanned
	rows[rowIndex].Set(false)

	// move to the next row
	rowIndex++
	if rowIndex >= numKeyboardMatrixRows {
		rowIndex = 0
	}
	s.row = rowIndex

	// energize the row we'll scan next
	rows[rowIndex].Set(true)

	return events
}

// scan one ADC per tick.
func (s *Scanner) scanADCs(events []Event) []Event {
	n := s.adc
	v := s.adcs[n].Get()

	events = append(events, NewADCEvent(uint8(n), v))

	n++
	if n >= len(s.adcs) {
		n = 0
	}
	s.adc = n

	return events
}

// scan the encoder once per tick.
func (s *Scanner) scanEncoder(events []Event) []Event {
	if delta := s.encoder.Read(); delta != 0 {
		events = append(events, NewEncoderEvent(0, int16(delta)))
	}

	return events
}
