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
	assert.Equal(t, Note(60).Pitch(), Pitch(0x50000000))  // middle C
	assert.Equal(t, Note(69).Pitch(), Pitch(0x5bfffffd))  // 440 Hz A
	assert.Equal(t, Note(71).Pitch(), Pitch(0x5eaaaaa7))  // B4
	assert.Equal(t, Note(72).Pitch(), Pitch(0x60000000))  // C5
	assert.Equal(t, Note(127).Pitch(), Pitch(0xa9555553)) // G10

	// Note(135) would be 20KHz if MIDI went that far.
	// We go that far...
	assert.Equal(t, Note(132).Pitch(), Pitch(0xb0000000)) // C11
	assert.Equal(t, Note(144).Pitch(), Pitch(0xc0000000)) // C12
	assert.Equal(t, Note(191).Pitch(), Pitch(0xfeaaaaa7)) // B15
	assert.Equal(t, Note(192).Pitch(), Pitch(0))          // not C16 (it wrapped)
}

// For each MIDI note, calculate its (scaled, int32) Frequency,
// then validate the result by converting back to a note number.
// The test passes if we get every MIDI note within two cents.
func TestPitchFrequency(t *testing.T) {
	// We can calculate frequencies beyond the MIDI range,
	// up to Note(138), which at 23.6KHz is a) ultrasonic,
	// and b) the last note before we hit our Nyquist rate.
	for noteIn := range 139 {
		pitch := Note(uint8(noteIn)).Pitch()
		noteOut := pitch.Frequency().Note()

		t.Logf("note: in %d => out %.2f\n", noteIn, noteOut)

		assert.Check(t, math.Abs(noteOut-float64(noteIn)) < 0.02)
	}
}

// Validate MinAudiblePitch.
func TestMinAudiblePitch(t *testing.T) {
	t.Log("lowest audible pitch =", MinAudiblePitch)
	minAudibleF := MinAudiblePitch.Frequency()
	minAudibleHz := minAudibleF.Hz()
	t.Logf("  => %v = %.3f hz", minAudibleF, minAudibleHz)
	assert.Check(t, minAudibleHz >= 20)

	var maxInfraP Pitch
	var maxInfraF Frequency
	for maxInfraP = MinAudiblePitch; maxInfraP > 0; maxInfraP-- {
		maxInfraF = maxInfraP.Frequency()
		if maxInfraF != minAudibleF {
			break
		}
	}
	t.Log("highest infrasound pitch =", maxInfraP)
	maxInfraHz := maxInfraF.Hz()
	t.Logf("  => %v = %.3f hz", maxInfraF, maxInfraHz)
	assert.Check(t, maxInfraHz < 20)
}

// Validate MaxAudiblePitch.
func TestMaxAudiblePitch(t *testing.T) {
	t.Log("highest audible pitch =", MaxAudiblePitch)
	maxAudibleF := MaxAudiblePitch.Frequency()
	maxAudibleHz := maxAudibleF.Hz()
	t.Logf("  => %v = %.3f hz", maxAudibleF, maxAudibleHz)
	assert.Check(t, maxAudibleHz <= 20_000)

	var minUltraP Pitch
	var minUltraF Frequency
	for minUltraP = MaxAudiblePitch; minUltraP < 0xffffffff; minUltraP++ {
		minUltraF = minUltraP.Frequency()
		if minUltraF != maxAudibleF {
			break
		}
	}
	t.Log("lowest ultrasound pitch =", minUltraP)
	minUltraHz := minUltraF.Hz()
	t.Logf("  => %v = %.3f hz", minUltraF, minUltraHz)
	assert.Check(t, minUltraHz > 20_000)
}
