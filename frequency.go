package picosynth

// A Frequency, stored as an uint32 scaled such that the highest
// representable value is [SampleRate] (± <1Hz rounding error.)
type Frequency uint32

// Common units of frequency.
const (
	KHz Frequency = (1000 << 32) / SampleRate
	Hz            = KHz / 1000
	BPM           = Hz / 60
)
