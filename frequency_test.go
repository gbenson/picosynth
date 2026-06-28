package picosynth

import (
	"math"
	"testing"

	"gotest.tools/v3/assert"
)

func TestPitchIntervals(t *testing.T) {
	assert.Equal(t, Octave, Pitch(0x10000000))
	assert.Equal(t, Semitone, Pitch(0x1555555))
	assert.Equal(t, Cent, Pitch(0x00369d0))
}

func TestNotePitch(t *testing.T) {
	assert.Equal(t, Note(0).Pitch(), Pitch(0))
	assert.Equal(t, Note(60).Pitch(), Pitch(0x50000000)) // middle C
	assert.Equal(t, Note(69).Pitch(), Pitch(0x5bfffffd)) // 440 Hz A
	assert.Equal(t, Note(71).Pitch(), Pitch(0x5eaaaaa7)) // B4
	assert.Equal(t, Note(72).Pitch(), Pitch(0x60000000)) // C5
	assert.Equal(t, Note(127).Pitch(), Pitch(0xa9555553))
}

// For each MIDI note, calculate its (scaled, int32) Frequency,
// then validate the result by converting back to a note number.
// The test passes if we get every MIDI note within two cents.
func TestPitchFrequency(t *testing.T) {
	for noteIn := range 128 {
		pitch := Note(uint8(noteIn)).Pitch()
		noteOut := noteFromFrequency(pitch.Frequency())

		t.Logf("note: in %d => out %.2f\n", noteIn, noteOut)

		assert.Check(t, math.Abs(noteOut-float64(noteIn)) < 0.02)
	}
}

func noteFromFrequency(f Frequency) float64 {
	freqHz := float64(f) * SampleRate / float64(1<<32)
	return math.Log2(freqHz/440)*12 + 69
}
