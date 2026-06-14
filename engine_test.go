package picosynth

import (
	"slices"
	"testing"
	"unsafe"

	"gotest.tools/v3/assert"
)

func generateAudio(ps *Engine, numSamples int) ([]int16, error) {
	// github.com/tinygo-org/pio/rp2-pio/piolib I2S devices
	// expect mono audio in []uint16 buffers...
	buf := make([]uint16, numSamples)
	if err := ps.Fill(buf); err != nil {
		return nil, err
	}

	// ...but the DAC interprets the values as signed.
	ptr := unsafe.Pointer(unsafe.SliceData(buf))
	return unsafe.Slice((*int16)(ptr), len(buf)), nil
}

func minmax(buf []int16) (lo, hi int16) {
	return slices.Min(buf), slices.Max(buf)
}

func TestSilence(t *testing.T) {
	ps := &Engine{}
	ps.init()
	ps.setVolume(MinVolume)
	ps.kt.notes[127] = Note(127)

	buf, err := generateAudio(ps, 128)
	assert.NilError(t, err)
	lo, hi := minmax(buf)
	t.Logf("min = %d, max = %d", lo, hi)
	assert.Check(t, lo >= int16(-1))
	assert.Check(t, hi <= int16(0))
}

func TestMinVolume(t *testing.T) {
	ps := &Engine{}
	ps.init()
	ps.setVolume(MinVolume + 1)
	ps.kt.notes[127] = Note(127)

	buf, err := generateAudio(ps, 128)
	assert.NilError(t, err)
	lo, hi := minmax(buf)
	t.Logf("min = %d, max = %d", lo, hi)
	assert.Check(t, lo >= int16(-64))
	assert.Check(t, lo < int16(-32))
	assert.Check(t, hi > int16(31))
	assert.Check(t, hi <= int16(63))
}

func TestMaxVolume(t *testing.T) {
	ps := &Engine{}
	ps.init()
	ps.setVolume(MaxVolume)
	ps.kt.notes[127] = Note(127)

	buf, err := generateAudio(ps, 128)
	assert.NilError(t, err)
	lo, hi := minmax(buf)
	t.Logf("min = %d, max = %d", lo, hi)
	assert.Check(t, lo < int16(-16384))
	assert.Check(t, hi > int16(16383))
}
