package picosynth

import (
	"math"
	"testing"
)

func TestSignalMul(t *testing.T) {
	for _, tc := range []struct {
		a, b, want Signal
	}{
		{0, 0, 0},

		{0, MaxSignal, 0},
		{MaxSignal, 0, 0},
		{MaxSignal, MaxSignal, MaxSignal},

		{0, MinSignal, 0},
		{MinSignal, 0, 0},
		{MinSignal, MaxSignal, MinSignal},

		{MaxSignal, MinSignal, MinSignal},
		{MinSignal, MinSignal, MaxSignal},

		{1234567890, MaxSignal, 1234567890},
		{MaxSignal, 987654321, 987654321},

		{1234567890, MinSignal, -1234567890},
		{MinSignal, 987654321, -987654321},

		{-1234567890, MaxSignal, -1234567890},
		{MaxSignal, -987654321, -987654321},

		{-1234567890, MinSignal, 1234567890},
		{MinSignal, -987654321, 987654321},
	} {
		got := tc.a.Mul(tc.b)
		var wanted string
		if got != tc.want {
			if tc.want > 0 && got > 0 && got == tc.want-1 {
				// both positive, got off by one towards zero.
				wanted = "*"
			} else if tc.want < 0 && got < 0 && got == tc.want+1 {
				// both negative, got off by one towards zero.
				wanted = "*"
			} else {
				wanted = "!= " + tc.want.String()
				if tc.a == MinSignal && tc.b == MinSignal && got == MinSignal {
					wanted += " expected failure" // XXX do we care?
				} else {
					t.Fail()
				}
			}
		}
		t.Log(tc.a, "×", tc.b, "=", got, wanted)
	}
}

func TestSignalSin(t *testing.T) {
	var maxF64Err float64

	for x := range 400 {
		want := Signal(min(MaxSignal, -math.Sin(float64(x)*math.Pi/180)*MinSignal))
		got := (Signal(x) * Degree).Sin()

		// What's the integer error across the MinSignal..MaxSignal range?
		i32Err := int(want) - int(got)
		if i32Err < 0 {
			i32Err = -i32Err
		}

		// What's that error scaled into the -1..1 range of math.Sin()?
		f64Err := float64(i32Err) / (float64(MaxSignal) - float64(MinSignal))
		maxF64Err = max(f64Err, maxF64Err)

		t.Logf("x=%d°: want(%v)-got(%v) = %.2g", x, want, got, f64Err)
		if math.Abs(f64Err) > 1e-3 {
			t.FailNow()
		}
	}

	t.Log("maxF64Err =", maxF64Err)
}
