package picosynth

import (
	"math"

	"gbenson.net/go/picosynth/internal/fmt"
)

// Signal is a signed Q1.31 fixed point number with range [-1,1).
//
// See https://en.wikipedia.org/wiki/Q_(number_format).
type Signal Int1_31

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

// Float64 returns the value as a float64 in the interval [-1,1).
func (x Signal) Float64() float64 {
	return float64(x) / -MinSignal
}

// Sin returns the sine of the signal x.
func (x Signal) Sin() Signal {
	return SineTable.Get(x)
}

// Mul returns the signed Q1.31 product of x and y.
//
// Note that MinSignal.Mul(MinSignal) overflows, returning MinSignal
// (-1) instead of the closer-to-correct MaxSignal (1).  This may or
// may not be corrected in future, so don't rely on it, but the point
// of using integer math over floating point here is for performance,
// not accuracy.
func (x Signal) Mul(y Signal) Signal {
	return Signal(x.Int1_31().Mul(y.Int1_31()))
}

// Int1_31 returns the signed Q1.31 representation of x.
func (x Signal) Int1_31() Int1_31 {
	return Int1_31(x)
}

// Int4_28 returns the signed Q4.28 representation of x.
// Three bits of precision are lost.
func (x Signal) Int4_28() Int4_28 {
	return x.Int1_31().Int4_28()
}

// String implements [fmt.Stringer].
func (v Signal) String() string {
	return fmt.Hex32Stringer("Signal", v)
}
