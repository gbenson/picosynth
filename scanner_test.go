package picosynth

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestConstants(t *testing.T) {
	// Specified
	assert.Equal(t, SampleRate, 48000)
	assert.Equal(t, MaxLatency, 10*time.Millisecond)
	assert.Equal(t, LongPressTimeout, 500*time.Millisecond)
	assert.Equal(t, ActivityTimeout, 30*time.Second)

	// Derived
	assert.Equal(t, BufferFrames, 32)
	assert.Equal(t, TickRate, 1500)
	assert.Equal(t, longPressTimeout, uint32(750))
	assert.Equal(t, activityTimeout, uint32(45000))
}
