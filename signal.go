package picosynth

import (
	"fmt"
	"math"
)

// Signal is a signed Q1.31 fixed point number with range [-1,1).
//
// See https://en.wikipedia.org/wiki/Q_(number_format).
type Signal int32

const (
	SignalBits = 32
	MinSignal  = math.MinInt32
	MaxSignal  = math.MaxInt32
)

// Constants for interpreting Signal as a phase with range [-π,π).
// Note that `Signal(Pi)` will overflow, but using Pi to construct
// fractions that resolve at compile time is e.g. Pi/4 is ok.
const (
	Pi     = -MinSignal
	Degree = Pi / 180
)

// Sin returns the sine of the signal x.
func (x Signal) Sin() Signal {
	return SineTable.Get(x)
}

// Mul returns the signed Q1.31 product of a and b.
//
// Note that MinSignal.Mul(MinSignal) overflows, returning MinSignal
// (-1) instead of the closer-to-correct MaxSignal (1).  This may or
// may not be corrected in future, so don't rely on it, but the point
// of using integer math over floating point here is for performance,
// not accuracy.
func (a Signal) Mul(b Signal) Signal {
	return Signal((int64(a) * int64(b)) >> 31)
}

// String implements [fmt.Stringer].
func (v Signal) String() string {
	return fmt.Sprintf("0x%08x", uint32(v))
}
