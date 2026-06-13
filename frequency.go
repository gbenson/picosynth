package picosynth

//go:generate go run make-frequency-table.go

// The frequency of a waveform, scaled such that the interval
// [0,MaxSignal+1] represents [0,SampleRate/2].  Note the upper
// bound will be inexact if SampleRate is not a power of two.
type Frequency Signal

// Common units of frequency.
const (
	KHz Frequency = (1000 << SignalBits) / SampleRate
	Hz            = KHz / 1000
	BPM           = Hz / 60
)

// The frequency of a waveform, stored logarithmically in a uint32.
// The upper 4 bits are the octave number with the rest being
// fractions of an octave.  In this form, transposition by musical
// intervals is addition and subtraction.
type Pitch uint32

// Musical intervals.
const (
	Octave   Pitch = 1 << 28
	Semitone       = Octave / 12
	Cent           = Octave / 1200
)

// Frequency returns the pitch as a Frequency.
func (p Pitch) Frequency() Frequency {
	// The most significant 4 bits of p are the octave number
	// (the MIDI note number divided by 12).
	const octaveShift = 28

	// The table spans one octave, the one with MIDI note numbers
	// [p2fTableOctave*12,p2fTableOctave*12+12), so the index we
	// need is the N most significant non-octave bits of p.
	const indexShift = octaveShift - log2p2fTableSize
	const indexMask = (1 << log2p2fTableSize) - 1

	// Extract the octave number and table index from p.
	octave := p >> octaveShift
	index := (p >> indexShift) & indexMask

	// Calculate the shift to transpose the table value to the
	// octave we need.
	resultShift := p2fTableOctave - octave

	// Lookup and transform the value.
	return Frequency(pitchToFrequencyTable[index] >> resultShift)
}

// A MIDI note, with Note(60) being middle C.
type Note uint8

const noteC4 = Note(60)

// NoNote is used to indicate "not a note".
const NoNote = Note(0xff)

// IsValid reports whether this is a valid MIDI note.
func (note Note) IsValid() bool {
	return note >= 0 && note < 128
}

// Pitch returns the note's center frequency as a Pitch.
func (note Note) Pitch() Pitch {
	octave := note / 12
	note -= octave * 12
	return Pitch(uint32(octave)<<28) | Pitch(note)*Semitone
}

// Transpose returns the note after transposition by the given interval.
func (note Note) Transpose(semitones int) Note {
	return Note(max(0, min(127, int(note)+semitones)))
}
