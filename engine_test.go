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

func TestSilence(t *testing.T) {
	ps := &Engine{}
	ps.volume = MinVolume
	ps.octave = 0
	ps.kt.Note = Note(127)

	buf, err := generateAudio(ps, 128)
	assert.NilError(t, err)
	assert.Check(t, slices.Min(buf) >= int16(-1))
	assert.Check(t, slices.Max(buf) <= int16(0))
}

func TestMinVolume(t *testing.T) {
	ps := &Engine{}
	ps.volume = MinVolume + 1
	ps.octave = 0
	ps.kt.Note = Note(127)

	buf, err := generateAudio(ps, 128)
	assert.NilError(t, err)
	assert.Check(t, slices.Min(buf) >= int16(-64))
	assert.Check(t, slices.Min(buf) < int16(-32))
	assert.Check(t, slices.Max(buf) > int16(31))
	assert.Check(t, slices.Max(buf) <= int16(63))
}

func TestMaxVolume(t *testing.T) {
	ps := &Engine{}
	ps.volume = MaxVolume
	ps.octave = 0
	ps.kt.Note = Note(127)

	buf, err := generateAudio(ps, 128)
	assert.NilError(t, err)
	assert.Check(t, slices.Min(buf) < int16(-16384))
	assert.Check(t, slices.Max(buf) > int16(16383))
}
