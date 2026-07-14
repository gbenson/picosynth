package picosynth

import (
	"slices"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"gbenson.net/go/picosynth/internal/adc"
)

func TestConstants(t *testing.T) {
	assert.Equal(t, SampleRate, 48000)
	assert.Equal(t, MaxLatency, 10*time.Millisecond)

	// Note we're very close to a step in buffer size; increasing
	// MaxLatency from 10ms to 10+2/3ms steps to 256-frame buffers.
	assert.Equal(t, BufferFrames, 128)

	// This is the rate of anything that happens once per Fill:
	// control surface scanning, voice pitch calculation, etc.
	assert.Equal(t, FillRate, 375)
}

func volumeTest(t *testing.T, volume int) (lo, hi int16) {
	t.Helper()

	ps := &Engine{}
	ps.init()
	ps.setVolume(volume)
	ps.kt.notes[127] = Note(127)
	ps.mem.Store(Filt1Mode, uint32(FilterNoFilter))

	buf := make([]int16, 128)
	assert.NilError(t, ps.Fill(buf))

	return slices.Min(buf), slices.Max(buf)
}

func TestSilence(t *testing.T) {
	lo, hi := volumeTest(t, MinVolume)
	t.Logf("min = %d, max = %d", lo, hi)
	assert.Check(t, lo >= int16(-1))
	assert.Check(t, hi <= int16(0))
}

func TestMinVolume(t *testing.T) {
	lo, hi := volumeTest(t, MinVolume+1)
	t.Logf("min = %d, max = %d", lo, hi)
	assert.Check(t, lo >= int16(-64))
	assert.Check(t, lo < int16(-32))
	assert.Check(t, hi > int16(31))
	assert.Check(t, hi <= int16(63))
}

func TestMaxVolume(t *testing.T) {
	lo, hi := volumeTest(t, MaxVolume)
	t.Logf("min = %d, max = %d", lo, hi)
	assert.Check(t, lo < int16(-16384))
	assert.Check(t, hi > int16(16383))
}

type MockADC uint16

func (a MockADC) Get() uint16 {
	return uint16(a)
}

// Create an Engine, apply the given ADC values, then generate a
// single sample and return the values that ended up set in the
// filter.
func filterTest(t *testing.T, cutADC, resADC uint16) (Frequency, Signal) {
	t.Helper()

	ps := &Engine{}
	ps.init()

	var adcs [2]adc.ADC
	adcs[0] = MockADC(cutADC)
	adcs[1] = MockADC(resADC)

	ps.pots = adcs[:]

	buf := make([]int16, 1)
	assert.NilError(t, ps.Fill(buf))

	return ps.filt1.Frequency, ps.filt1.Resonance
}

func TestMinCutoff(t *testing.T) {
	cut, _ := filterTest(t, 0x0000, 0x1234)
	assert.Check(t, Within01Percent(cut.Hz(), 20))
}

func TestMaxCutoff(t *testing.T) {
	cut, _ := filterTest(t, 0xffff, 0x1234)
	assert.Check(t, Within01Percent(cut.Hz(), 20_000))
}

func TestMinResonance(t *testing.T) {
	_, res := filterTest(t, 0x1234, 0x0000)
	assert.Equal(t, res, Signal(0))
}

func TestMaxResonance(t *testing.T) {
	_, res := filterTest(t, 0x1234, 0xffff)
	assert.Check(t, Within001Percent(res.Float64(), 1))
}
