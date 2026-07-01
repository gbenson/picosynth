package picosynth

import (
	"slices"
	"testing"

	"gotest.tools/v3/assert"
)

func volumeTest(t *testing.T, volume int) (lo, hi int16) {
	t.Helper()

	ps := &Engine{}
	ps.init()
	ps.setVolume(volume)
	ps.kt.notes[127] = Note(127)
	ps.storeFilterMode(Filt1Mode, FilterNoFilter)

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
