package picosynth

type PhaseAccumulator struct {
	Frequency Frequency
	Phase     uint32
}

func (pa *PhaseAccumulator) Step() {
	pa.Phase += uint32(pa.Frequency)
}
