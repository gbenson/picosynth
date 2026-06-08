package picosynth

type Shaper func(phase Signal, shape Signal) Signal

// SineShaper is a [Shaper] that generates sine waves, with the shape
// parameter offsetting input phase.  With shape 0 it will generate a
// sine wave; with shape Pi/2 a cosine wave.
func SineShaper(phase Signal, shape Signal) Signal {
	return (phase + shape).Sin()
}

// TriSawShaper is a [Shaper] that generates combinations of triangle
// and saw waves.  With shape 0 it generates a triangle wave with the
// same phase as a sine wave: positive in the first half-cycle and
// negative in the second.  With shape MinSignal it generates a
// falling sawtooth waveform: again, positive in the first half-cycle
// and negative in the second.  With shape MaxSignal it generates a
// rising sawtooth waveform: negative in the first half-cycle,
// positive in the second.
func TriSawShaper(phase Signal, shape Signal) Signal {
	if shape == MaxSignal {
		return phase
	}
	panic("TriSawShaper: not implemented")
}
