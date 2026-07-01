package picosynth

// modeled after golang.org/x/image/math/fixed

// Int1_31 is a signed Q1.31 fixed-point number with range [-1,1).
//
// The integer part ranges from -1 to 0, inclusive.
// The fractional part has 31 bits of precision.
type Int1_31 int32

// Int2_30 is a signed Q2.30 fixed-point number with range [-2,2).
//
// The integer part ranges from -2 to 0, inclusive.
// The fractional part has 30 bits of precision.
type Int2_30 int32

// Int4_28 is a signed Q4.28 fixed-point number with range [-8,8).
//
// The integer part ranges from -8 to 7, inclusive.
// The fractional part has 28 bits of precision.
type Int4_28 int32

// Conversions ----------------------------------------------------

// Int4_28 returns the Int1_31 value x as an Int4_28.
// Three bits of precision are lost.
func (x Int1_31) Int4_28() Int4_28 {
	return Int4_28(x >> 3)
}

// Int4_28 returns the Int2_30 value x as an Int4_28.
// Two bits of precision are lost.
func (x Int2_30) Int4_28() Int4_28 {
	return Int4_28(x >> 2)
}

// Multiplication -------------------------------------------------

// Mul returns x*y in Q1.31 fixed-point arithmetic.
func (x Int1_31) Mul(y Int1_31) Int1_31 {
	return Int1_31((int64(x)*int64(y) + 1<<30) >> 31)
}

// Mul returns x*y in Q4.28 fixed-point arithmetic.
func (x Int4_28) Mul(y Int4_28) Int4_28 {
	return Int4_28((int64(x)*int64(y) + 1<<27) >> 28)
}

// Clamped multiplication -----------------------------------------

// ClampedMul returns x*y in Q4.28 fixed-point arithmetic.
func (x Int4_28) ClampedMul(y Int4_28) Int4_28 {
	z := (int64(x)*int64(y) + 1<<27) >> 28
	switch {
	case z < MinSignal:
		return MinSignal
	case z > MaxSignal:
		return MinSignal
	default:
		return Int4_28(z)
	}
}
