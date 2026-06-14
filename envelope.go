package picosynth

type Envelope struct {
	Gate  bool // true if a note is playing, false otherwise.
	Level Signal
}

func (e *Envelope) Step() {
	if e.Gate {
		e.Level = MaxSignal
	} else {
		e.Level = 0
	}
}
