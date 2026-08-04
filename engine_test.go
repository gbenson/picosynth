package picosynth

import (
	"slices"
	"sync/atomic"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"gbenson.net/go/picosynth/internal/hw/drivers/encoders"
	"gbenson.net/go/picosynth/internal/hw/machine"
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

	assert.Equal(t, LongPressTimeout, 500*time.Millisecond)
	assert.Equal(t, longPressTimeout, uint32(187))

	assert.Equal(t, ActivityTimeout, 30*time.Second)
	assert.Equal(t, activityTimeout, uint32(11250))
}

func volumeTest(t *testing.T, volume int) (lo, hi int16) {
	t.Helper()
	WithTestEmulator(t, &fillTestEmulator{t: t})

	ps := &Picosynth{}
	assert.NilError(t, ps.Init())
	ps.SetVolume(volume)
	ps.keytracker.notes[127] = Note(127)
	ps.Engine.Memory.Store(Filt1Mode, uint32(FilterNoFilter))

	buf := make([]int16, 128)
	assert.NilError(t, ps.fill(buf))

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

// Create an Engine, apply the given ADC values, then generate a
// single sample and return the values that ended up set in the
// filter.
func filterTest(t *testing.T, cutADC, resADC uint16) (Frequency, Signal) {
	t.Helper()
	e := WithTestEmulator(t, &fillTestEmulator{t: t})

	ps := &Picosynth{}
	assert.NilError(t, ps.Init())

	e.setADC(ps.pots[0], cutADC)
	e.setADC(ps.pots[1], resADC)

	buf := make([]int16, 1)
	assert.NilError(t, ps.fill(buf))

	return ps.Engine.filt1.Frequency, ps.Engine.filt1.Resonance
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

type fillTestEmulator struct {
	t *testing.T

	adcInitialized atomic.Bool
	adcConfigured  [4]atomic.Bool
	adcValue       [4]atomic.Uint32

	encoderConfigured atomic.Bool
	encoderPosition   atomic.Int32
}

func (e *fillTestEmulator) InitADC() {
	t := e.t

	assert.Equal(t, e.adcInitialized.Swap(true), false, "already initialized")
}

func (e *fillTestEmulator) adcIndex(a machine.ADC) int {
	t := e.t

	switch a.Pin {
	case machine.ADC0:
		return 0
	case machine.ADC1:
		return 1
	case machine.ADC2:
		return 2
	case machine.ADC3:
		return 3
	}

	t.Logf("invalid ADC pin GPIO%d", int(a.Pin))
	t.FailNow()
	panic("should not reach here")
}

func (e *fillTestEmulator) ConfigureADC(a machine.ADC, c machine.ADCConfig) error {
	t := e.t
	n := e.adcIndex(a)

	assert.Equal(t, c, machine.ADCConfig{})
	assert.Check(t, e.adcConfigured[n].Swap(true) == false, "already configured")

	return nil
}

func (e *fillTestEmulator) setADC(a machine.ADC, v uint16) {
	e.adcValue[e.adcIndex(a)].Store(uint32(v))
}

func (e *fillTestEmulator) GetADC(a machine.ADC) uint16 {
	t := e.t
	n := e.adcIndex(a)

	assert.Check(t, e.adcConfigured[n].Load(), "not configured")

	return uint16(e.adcValue[n].Load())
}

func (e *fillTestEmulator) ConfigureEncoder(
	enc encoders.QuadratureDevice,
	config encoders.QuadratureConfig,
) error {
	t := e.t

	assert.Equal(t, enc.PinA, machine.GP0)
	assert.Equal(t, enc.PinB, machine.GP1)
	assert.Equal(t, config, encoders.QuadratureConfig{Precision: 4})

	assert.Check(t, e.encoderConfigured.Swap(true) == false, "already configured")

	return nil
}

func (e *fillTestEmulator) EncoderPosition(enc encoders.QuadratureDevice) int {
	t := e.t

	assert.Check(t, e.encoderConfigured.Load(), "not configured")

	return int(e.encoderPosition.Load())
}
