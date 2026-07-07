package picosynth

import (
	"fmt"
	"math"
	"os"
	"slices"
	"strings"
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

// stretch4to12 maps 0=>0, 1=>0x111, 2=>0x222, ... 15=>0xfff.
func stretch4to12(v int) int {
	return (v << 8) | (v << 4) | v
}

func stretch8to12(v int) int {
	return (v << 4) | (v >> 4)
}

func dontTestSawtoothResponse2(t *testing.T) {
	const NumCutADCSteps = 256
	const NumResADCSteps = 16

	// SampleRate/4 yields essentially the same results as SampleRate,
	// except it's not obvious what's happening with the resonant low
	// notes.
	const NumSamples = SampleRate // / 4
	var buf [5][NumSamples]Signal

	I := buf[0][:]
	L := buf[1][:]
	H := buf[2][:]
	B := buf[3][:]
	N := buf[4][:]

	outputs := [][]Signal{L, H, B, N}

	// 12 × 16 × 16 = 3072 evaluations
	for note := Note(5); note < 128; note += 11 {
		oscFreq := note.Pitch().Frequency()
		//oscHz := int(oscFreq.Hz())

		var absMin [NumCutADCSteps][NumResADCSteps]Signal
		var absMax [NumCutADCSteps][NumResADCSteps]Signal
		var tailMin [NumCutADCSteps][NumResADCSteps]Signal
		var tailMax [NumCutADCSteps][NumResADCSteps]Signal
		var rmsMax [NumCutADCSteps][NumResADCSteps]Signal

		for ci := range NumCutADCSteps {
			cutADC := stretch8to12(ci) << 4
			cutPitch := Pitch(cutADC)
			cutPitch *= (MaxAudiblePitch - MinAudiblePitch) >> 16
			cutPitch += MinAudiblePitch
			cut := cutPitch.Frequency()
			//cutHz := int(cut.Hz())

			for ri := range NumResADCSteps {
				resADC := stretch4to12(ri) << 4
				res := Signal(resADC << 15)
				//resPercent := int(100 * res.Float64())

				osc := PhaseAccumulator{Frequency: oscFreq}
				f := ChamberlinFilter{Frequency: cut, Resonance: res}

				for i := range NumSamples {
					osc.Step()
					f.Input = osc.Phase
					f.Step()

					I[i] = f.Input
					L[i] = f.Lout
					H[i] = f.Hout
					B[i] = f.Bout
					N[i] = f.Nout
				}

				// could dump all series here if required

				for _, series := range outputs {
					vmin := slices.Min(series)
					vmax := slices.Max(series)
					//t.Logf("%03d:0x%04x:0x%04x:%d: [0x%08x,0x%08x]",
					//	note, cutADC>>4, resADC>>4, i, uint32(vmin), vmax)

					// record extrema for stability and headroom checks
					absMin[ci][ri] = min(absMin[ci][ri], vmin)
					absMax[ci][ri] = max(absMax[ci][ri], vmax)

					// record tail rms and extrema for setting compression
					const TailSize = SampleRate / 8
					tail := series[NumSamples-TailSize:]
					tmin := slices.Min(tail)
					tmax := slices.Max(tail)

					tailMin[ci][ri] = min(tailMin[ci][ri], tmin)
					tailMax[ci][ri] = max(tailMax[ci][ri], tmax)

					var sumSquares int64
					for _, v := range tail {
						sumSquares += int64(v) * int64(v)
					}
					meanSquares := float64(sumSquares) / float64(TailSize)
					vrms := Signal(math.Sqrt(meanSquares))

					rmsMax[ci][ri] = max(rmsMax[ci][ri], vrms)
				}
			}
		}

		filename := fmt.Sprintf("note%03d-tailminmax.gp", note)
		t.Log("writing", filename)
		fp, err := os.Create(filename)
		assert.NilError(t, err)
		defer fp.Close()
		for ci := range NumCutADCSteps {
			cutADC := stretch8to12(ci) << 4
			cutPitch := Pitch(cutADC)
			cutPitch *= (MaxAudiblePitch - MinAudiblePitch) >> 16
			cutPitch += MinAudiblePitch
			cut := cutPitch.Frequency()

			var fields [NumResADCSteps + 1]string
			fields[0] = fmt.Sprint(cut.Hz())

			for ri := range NumResADCSteps {
				fields[ri+1] = fmt.Sprint(
					max(-(tailMin[ci][ri].Float64()),
						tailMax[ci][ri].Float64()) * 8)
			}
			fmt.Fprintln(fp, strings.Join(fields[:], "\t"))
		}
	}

	t.Fail()
}

func dontTestSawtoothResponse1(t *testing.T) {
	const NumCutADCSteps = 16
	const NumResADCSteps = 16

	var absMin [NumCutADCSteps][NumResADCSteps]Signal
	var absMax [NumCutADCSteps][NumResADCSteps]Signal
	var tailMin [NumCutADCSteps][NumResADCSteps]Signal
	var tailMax [NumCutADCSteps][NumResADCSteps]Signal
	var rmsMax [NumCutADCSteps][NumResADCSteps]Signal

	// SampleRate/4 yields essentially the same results as SampleRate,
	// except it's not obvious what's happening with the resonant low
	// notes.
	const NumSamples = SampleRate // / 4
	var buf [5][NumSamples]Signal

	I := buf[0][:]
	L := buf[1][:]
	H := buf[2][:]
	B := buf[3][:]
	N := buf[4][:]

	outputs := [][]Signal{L, H, B, N}

	// 12 × 16 × 16 = 3072 evaluations
	for note := Note(5); note < 128; note += 11 {
		oscFreq := note.Pitch().Frequency()
		//oscHz := int(oscFreq.Hz())

		for ci := range NumCutADCSteps {
			cutADC := stretch4to12(ci) << 4
			cutPitch := Pitch(cutADC)
			cutPitch *= (MaxAudiblePitch - MinAudiblePitch) >> 16
			cutPitch += MinAudiblePitch
			cut := cutPitch.Frequency()
			//cutHz := int(cut.Hz())

			for ri := range NumResADCSteps {
				resADC := stretch4to12(ri) << 4
				res := Signal(resADC << 15)
				//resPercent := int(100 * res.Float64())

				osc := PhaseAccumulator{Frequency: oscFreq}
				f := ChamberlinFilter{Frequency: cut, Resonance: res}

				for i := range NumSamples {
					osc.Step()
					f.Input = osc.Phase
					f.Step()

					I[i] = f.Input
					L[i] = f.Lout
					H[i] = f.Hout
					B[i] = f.Bout
					N[i] = f.Nout
				}

				// could dump all series here if required

				for _, series := range outputs {
					vmin := slices.Min(series)
					vmax := slices.Max(series)
					//t.Logf("%03d:0x%04x:0x%04x:%d: [0x%08x,0x%08x]",
					//	note, cutADC>>4, resADC>>4, i, uint32(vmin), vmax)

					// record extrema for stability and headroom checks
					absMin[ci][ri] = min(absMin[ci][ri], vmin)
					absMax[ci][ri] = max(absMax[ci][ri], vmax)

					// record tail rms and extrema for setting compression
					const TailSize = SampleRate / 8
					tail := series[NumSamples-TailSize:]
					tmin := slices.Min(tail)
					tmax := slices.Max(tail)

					if tmin > -0x70000000 && tmax < 0x70000000 {
						tailMin[ci][ri] = min(tailMin[ci][ri], tmin)
						tailMax[ci][ri] = max(tailMax[ci][ri], tmax)
					} // else presumed unstable

					var sumSquares int64
					for _, v := range tail {
						sumSquares += int64(v) * int64(v)
					}
					meanSquares := float64(sumSquares) / float64(TailSize)
					vrms := Signal(math.Sqrt(meanSquares))

					rmsMax[ci][ri] = max(rmsMax[ci][ri], vrms)
				}

				// // dump the most unstable series
				// if unstable < 8 {
				// 	continue
				// }

				// filename := fmt.Sprintf(
				// 	"note%03d-cut0x%04x-res0x%04x-series.gp",
				// 	note, cutADC, resADC,
				// )
				// t.Log("writing", filename)
				// fp, err := os.Create(filename)
				// assert.NilError(t, err)
				// defer fp.Close()
				// fmt.Fprintln(fp, "# Time\tInput\tLout\tHout\tBout\tNout")
				// for i := range NumSamples {
				// 	fmt.Fprintf(
				// 		fp,
				// 		"%g\t%g\t%g\t%g\t%g\t%g\n",
				// 		float64(i)/SampleRate,
				// 		I[i].Float64(),
				// 		L[i].Float64()*8,
				// 		H[i].Float64()*8,
				// 		B[i].Float64()*8,
				// 		N[i].Float64()*8,
				// 	)
				// }
			}
		}
	}

	// collate min and max rms and peak for the stable region.
	var maxRMS, minPeak, maxPeak, minTail, maxTail Signal
	var meanRMSq, meanRMSr int64
	for ci := range NumCutADCSteps {
		for ri := range NumResADCSteps {
			thisMinP := absMin[ci][ri]
			if thisMinP < -0x70000000 {
				continue // unstable
			}

			thisMaxP := absMax[ci][ri]
			if thisMaxP > 0x70000000 {
				continue // unstable
			}

			minPeak = min(minPeak, thisMinP)
			maxPeak = max(maxPeak, thisMaxP)

			minTail = min(minTail, tailMin[ci][ri])
			maxTail = max(maxTail, tailMax[ci][ri])

			thisRMS := rmsMax[ci][ri]

			maxRMS = max(maxRMS, thisRMS)
			meanRMSq += int64(maxRMS)
			meanRMSr += 1
		}
	}

	meanRMS := Signal(meanRMSq / meanRMSr)

	t.Log("stable region:")
	t.Log(" - extrema:")
	t.Logf("    - minimum: -0x%08x", -int32(minPeak))
	t.Logf("    - maximum:  0x%08x", int32(maxPeak))
	t.Log(" - tail:")
	t.Logf("    - minimum: -0x%08x", -int32(minTail))
	t.Logf("    - maximum:  0x%08x", int32(maxTail))
	t.Log("    - rms:")
	t.Logf("       - mean: 0x%08x", int32(meanRMS))
	t.Logf("       - peak: 0x%08x", int32(maxRMS))

	// With 1 second series (NumSamples = SampleRate)
	//
	// chamberlin-filter_test.go:209: writing note05-cut0x1110-res0xfff0-series.gp
	// chamberlin-filter_test.go:209: writing note27-cut0x8880-res0xbbb0-series.gp
	// chamberlin-filter_test.go:209: writing note49-cut0x4440-res0xfff0-series.gp
	// chamberlin-filter_test.go:264: stable region:
	// chamberlin-filter_test.go:265:  - extrema:
	// chamberlin-filter_test.go:266:     - minimum: -0x6e7bf166
	// chamberlin-filter_test.go:267:     - maximum:  0x6e38cf1e
	// chamberlin-filter_test.go:268:  - tail:
	// chamberlin-filter_test.go:269:     - minimum: -0x6329c258
	// chamberlin-filter_test.go:270:     - maximum:  0x64f80e11
	// chamberlin-filter_test.go:271:     - rms:
	// chamberlin-filter_test.go:272:        - mean: 0x0255535c
	// chamberlin-filter_test.go:273:        - peak: 0x0255f785

	// With 0.25 second series (NumSamples = SampleRate/4)
	// chamberlin-filter_test.go:211: writing note05-cut0x6660-res0xfff0-series.gp
	// chamberlin-filter_test.go:211: writing note49-cut0x4440-res0xfff0-series.gp
	// chamberlin-filter_test.go:211: writing note60-cut0x5550-res0xeee0-series.gp
	// chamberlin-filter_test.go:266: stable region:
	// chamberlin-filter_test.go:267:  - extrema:
	// chamberlin-filter_test.go:268:     - minimum: -0x6ad8970d
	// chamberlin-filter_test.go:269:     - maximum:  0x6e38cf1e
	// chamberlin-filter_test.go:270:  - tail:
	// chamberlin-filter_test.go:271:     - minimum: -0x6ad8970d
	// chamberlin-filter_test.go:272:     - maximum:  0x6cb0c583
	// chamberlin-filter_test.go:273:     - rms:
	// chamberlin-filter_test.go:274:        - mean: 0x025599f7
	// chamberlin-filter_test.go:275:        - peak: 0x025641ce

	// note49-cut0x4440-res0xfff0-series.gp is resonating hard,
	// its Bout exceeds ±6!
	//
	// note60-cut0x5550-res0xeee0-series.gp Bout hits nearly +2;
	//   its Hout and Bout are ±1.5 so probably high rms,
	//   though it got logged because its Nout has the highest
	//   rms of all... it's letting its entire input straight through?
	//
	// the above two are likely good limiting-case testcases;
	//  note05-cut0x6660-res0xfff0-series.gp is also resonating,
	//  with slightly higher ampitide than the note49 one even
	//  (maybe) but it's much slower so you need a full second
	//  of samples to see it.

	// plot a heat map of tail peak
	colors := []int{
		/* 0 */ 16, // almost black -> no stable values
		/* 1 */ 18, // navy
		/* 2 */ 21, // mid blue <- most are here
		/* 3 */ 229, // lemon
		/* 4 */ 220, // orange-yellow
		/* 5 */ 208, // orange
		/* 6 */ 196, // red
	}

	t.Log(" RES")
	for ri := range NumResADCSteps {
		resADC := stretch4to12(ri) << 4
		res := Signal(resADC << 15)

		row := fmt.Sprintf("%4.1f ", res.Float64())

		for ci := range NumCutADCSteps {
			minTail := tailMin[ci][ri]
			maxTail := tailMax[ci][ri]

			peak := max(int64(maxTail), -int64(minTail))
			row = fmt.Sprintf("%s\x1B[48;5;%dm  ", row, colors[peak>>28])
		}
		t.Logf("%s\x1B[0m", row)
	}
	t.Log()

	t.Fail()
}

func dontTestSawtoothStability(t *testing.T) {
	const NumSamples = SampleRate
	var buf [5][NumSamples]Signal

	I := buf[0][:]
	L := buf[1][:]
	H := buf[2][:]
	B := buf[3][:]
	N := buf[4][:]

	outputs := [][]Signal{L, H, B, N}

	for _, tc := range []struct {
		note   Note
		cutADC uint16
		resADC uint16
	}{
		{60, 0xddd0, 0x1110},
	} {
		oscFreq := tc.note.Pitch().Frequency()
		//oscHz := int(oscFreq.Hz())

		cutPitch := Pitch(tc.cutADC)
		cutPitch *= (MaxAudiblePitch - MinAudiblePitch) >> 16
		cutPitch += MinAudiblePitch
		cut := cutPitch.Frequency()
		//cutHz := int(cut.Hz())

		res := Signal(tc.resADC << 15)
		//resPercent := int(100 * res.Float64())

		osc := PhaseAccumulator{Frequency: oscFreq}
		f := ChamberlinFilter{Frequency: cut, Resonance: res}

		for i := range NumSamples {
			osc.Step()
			f.Input = osc.Phase
			f.Step()

			I[i] = f.Input
			L[i] = f.Lout
			H[i] = f.Hout
			B[i] = f.Bout
			N[i] = f.Nout
		}

		// dump the time series
		filename := fmt.Sprintf(
			"note%03d-cut0x%04x-res0x%04x-series.gp",
			tc.note, tc.cutADC, tc.resADC,
		)
		t.Log("writing", filename)
		fp, err := os.Create(filename)
		assert.NilError(t, err)
		defer fp.Close()
		fmt.Fprintln(fp, "# Time\tInput\tLout\tHout\tBout\tNout")
		for i := range NumSamples {
			fmt.Fprintf(
				fp,
				"%g\t%g\t%g\t%g\t%g\t%g\n",
				float64(i)/SampleRate,
				I[i].Float64(),
				L[i].Float64()*8,
				H[i].Float64()*8,
				B[i].Float64()*8,
				N[i].Float64()*8,
			)
		}

		// If it runs away and starts wrapping you get static,
		// which we detect by counting turning points; e.g. for
		// note060-cut0xddd0-res0x1110-series.gp wrapping we see:
		//
		//  - L: turning points: 36115 of: 48000
		//  - H: turning points: 36126 of: 48000
		//  - B: turning points: 36028 of: 48000
		//  - N: turning points: 35661 of: 48000
		for _, series := range outputs {
			var turningPoints int
			var lastv Signal
			var lastIncreasing bool
			for _, v := range series {
				if increasing := v > lastv; increasing != lastIncreasing {
					turningPoints++
					lastIncreasing = increasing
				}
				lastv = v
			}
			t.Log("turning points:", turningPoints, "of:", len(series))
		}
	}
	t.Fail()
}
