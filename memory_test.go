package picosynth

import (
	"math"
	"testing"

	"gotest.tools/v3/assert"
)

func TestEditable(t *testing.T) {
	var r Register
	for i := range NumRegisters {
		r.n = i
		switch {
		case i < 0x80:
			assert.Check(t, r.Editable())
		case i < 0x100:
			assert.Check(t, !r.Editable())
		default:
			assert.Check(t, i < 512)
			assert.Check(t, r.Editable())
		}
	}

	for i, want := range map[int]bool{
		0x000: true,  // no name
		0x040: true,  // Osc1Pitch (bias)
		0x043: true,  // no name
		0x044: true,  // Osc2Pitch
		0x048: true,  // no name
		0x080: false, // VoicePitch
		0x081: false, // no name
		0x0c4: false, // ModulatedOsc2Pitch
		0x0c5: false, // no name
		0x100: true,  // Osc1PitchSrc0
		0x10c: true,  // no name
		0x180: true,  // Env1AttSrc0
	} {
		r.n = i
		assert.Equal(t, r.Editable(), want)
	}
}

func TestMatrixOutputCalculation(t *testing.T) {
	var ps Engine
	ps.init()
	m := &ps.Memory

	out := m.Register(ModulatedLFO2Rate) // value used by LFO2 for this step
	in1 := m.Register(LFO2Rate)          // bias value (from LFO2 rate knob)
	in2 := m.Register(VoicePitch)
	rc2 := m.Register(LFO2RateSrc2) // cell 2 of LFORate matrix row
	op1 := m.Register(ModulatedOsc1Pitch)
	op2 := m.Register(ModulatedOsc2Pitch)

	for _ = range 2 {
		assert.Equal(t, in1.Load(), uint32(0))
		assert.Equal(t, in2.Load(), uint32(0))
		assert.Equal(t, out.Load(), uint32(0))
		assert.Equal(t, rc2.Load(), uint32(0))
		assert.Equal(t, op1.Load(), uint32(0))
		assert.Equal(t, op2.Load(), uint32(0))

		m.Step()
	}

	in1.Store(uint32(120 * BPM))
	assert.Equal(t, in1.Load(), uint32(0x0002bae8))
	assert.Equal(t, out.Load(), uint32(0x0002bae7)) // SO close!

	in1.Store(uint32(135 * BPM))
	assert.Equal(t, in1.Load(), uint32(0x00031245))
	assert.Equal(t, out.Load(), uint32(0x0002bae7)) // not updated
	m.Step()
	assert.Equal(t, out.Load(), uint32(0x00031244)) // updated this time

	in1.Store(uint32(120 * BPM))
	in2.Store(uint32(noteC4.Pitch()))
	wantNote := float64(noteC4)
	assert.Equal(t, in2.Load(), uint32(0x50000000))
	assert.Equal(t, out.Load(), uint32(0x00031244)) // not updated
	m.Step()
	assert.Equal(t, out.Load(), uint32(0x0002bae7)) // updated now but no in2

	cc := (uint32(VoicePitch) << 24) | 0x123456
	rc2.Store(uint32(cc))
	assert.Equal(t, rc2.Load(), uint32(cc))
	assert.Equal(t, out.Load(), uint32(0x0002bae7)) // not updated
	m.Step()
	assert.Equal(t, out.Load(), uint32(0x0b6370a7)) // updated now

	// bonus: check ModulatedOsc[12]Pitch are following VoicePitch
	assert.Equal(t, op1.Load(), uint32(0x4fffff60))
	assert.Equal(t, op2.Load(), uint32(0x4fffff60))

	// side note: how close is 0x4fffff60 to 0x50000000?
	wantPitch := Pitch(in2.Load())
	gotPitch := Pitch(op1.Load())
	assert.Equal(t, wantPitch, Pitch(0x50000000))
	assert.Equal(t, gotPitch, Pitch(0x4fffff60))

	gotNote := noteFromFrequency(gotPitch.Frequency())

	t.Logf("want: MIDI %v, got MIDI %v", wantNote, gotNote)

	// Two cents is the standard we validate Pitch.Frequency() against.
	assert.Check(t, math.Abs(gotNote-wantNote) < 0.02)
}
