package picosynth

import (
	"math"
	"testing"

	"gotest.tools/v3/assert"
)

func TestPhaseAccumulator(t *testing.T) {
	var pa PhaseAccumulator

	// Check the initial values.
	assert.Equal(t, pa.Frequency, Frequency(0))
	assert.Equal(t, pa.Phase, Signal(0))

	// Stepping at 0 Hz shouldn't do anything.
	for _ = range SampleRate {
		pa.Step()
		assert.Equal(t, pa.Phase, Signal(0))
	}

	// Stepping at SampleRate/N should wrap every N timesteps.
	t.Logf("SampleRate: %g KHz", float64(SampleRate)/1000)
	for n := range 100 {
		if n < 2 {
			continue
		}

		pa.Frequency = Frequency(0xffffffff / n)
		t.Logf("SampleRate/%d => Frequency(0x%08x)", n, pa.Frequency)

		// run for 1s, counting positive-negative transitions
		wraps := 0
		pa.Phase = 0

		for _ = range SampleRate {
			wasHigh := pa.Phase >= 0
			pa.Step()
			nowLow := pa.Phase < 0
			if wasHigh && nowLow {
				wraps++
			}
		}

		stepsPerWrap := int(math.Round(float64(SampleRate) / float64(wraps)))
		t.Logf("%d wraps in %d steps => %d steps/wrap", wraps, SampleRate, stepsPerWrap)
		assert.Equal(t, stepsPerWrap, n)
	}
}
