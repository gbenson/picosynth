package picosynth

import (
	"slices"
	"testing"

	"gotest.tools/v3/assert"
)

func TestStepResponse(t *testing.T) {
	var absMin, absMax int32

	for ti, tc := range []struct {
		F, Q1 float64
	}{
		// self-oscillates when calculated with float64, with one
		// output peaking at 4.73205.
		{20000, 0},
		// settles within 500 samples when calculated with float64,
		// but one output peaked at 4.8291!
		{17027.8, 0.214497},
		// the highest peak I ever saw (4.88903!)
		{20000, 0.00626351},
		// most frequencies peaked at or near 1 with low resonance.
		{20.019, 1},
		{43.701, 1.99418},
		{299.658, 1.88208},
		{3870.88, 2},
		{7090.18, 0.841689},
		{10361.7, 0.447203},
		{13166, 0},
	} {
		f := ChamberlinFilter{
			Frequency: Frequency(tc.F * (1 << 32) / SampleRate),
			Resonance: Signal((2 - tc.Q1) * MaxSignal / 2),
			Input:     MaxSignal, // step
		}

		t.Logf("F: %.3fHz => %v => %.3fHz",
			tc.F, f.Frequency, f.Frequency.Hz())
		assert.Check(t, NearlyEqual(tc.F, f.Frequency.Hz())) // sanity

		t.Logf("Q1: %.6f => Resonance: %.2f%% %v",
			tc.Q1, 100*float64(f.Resonance)/MaxSignal, f.Resonance)

		// The slowest testcase (F=20.019Hz,Q1=1) takes ~0.2s to
		// stabilize, so we use SampleRate/4 = 0.25s for leeway.
		const NumSamples = SampleRate / 4
		var buf [4][NumSamples]Signal

		L := buf[0][:]
		H := buf[1][:]
		B := buf[2][:]
		N := buf[3][:]

		for i := range NumSamples {
			f.Step()

			L[i] = f.Lout
			H[i] = f.Hout
			B[i] = f.Bout
			N[i] = f.Nout
		}

		// fn := fmt.Sprintf("step-response-%d.gp", ti)
		// t.Log("H(s) saved as:", fn)
		// fp, err := os.Create(fn)
		// assert.NilError(t, err)
		// defer fp.Close()
		// fmt.Fprintln(fp, "# Time\tLP\tHP\tBP\tBR")
		// for i := range NumSamples {
		// 	fmt.Fprintf(
		// 		fp,
		// 		"%g\t%g\t%g\t%g\t%g\n",
		// 		float64(i)/SampleRate,
		// 		L[i].Float64()*8,
		// 		H[i].Float64()*8,
		// 		B[i].Float64()*8,
		// 		N[i].Float64()*8,
		// 	)
		// }

		for i, ts := range [][]Signal{L, H, B, N} {
			vmin := int32(slices.Min(ts))
			vmax := int32(slices.Max(ts))
			t.Logf("%d: %d: [0x%08x,0x%08x]", ti, i, vmin, vmax)

			// Check we have plenty of headroom (ie nothing clipped).
			assert.Check(t, vmin > -0x50000000)
			assert.Check(t, vmax < 0x50000000)

			absMin = min(absMin, vmin)
			absMax = max(absMax, vmax)
		}

		// XXX check other stuff?
		//  - if t.Q1 != 0, did all outputs settle?
		//  - if t.Q1 == 0, is it self-oscillating?
	}

	// Check at least one signal would have clipped with less range.
	assert.Check(t, absMin < -0x40000000 || absMax >= 0x40000000)
}
