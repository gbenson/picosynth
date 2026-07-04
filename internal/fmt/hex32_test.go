package fmt

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestHex32(t *testing.T) {
	assert.Equal(t, Hex32(14), "0000000e")
	assert.Equal(t, Hex32(0xcafebabe), "cafebabe")
	assert.Equal(t, Hex32(int32(-12345678)), "ff439eb2")
	assert.Equal(t, Hex32(uint32(0xcdef1234)), "cdef1234")
}
