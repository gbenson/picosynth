package picosynth

import "testing"

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
