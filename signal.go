package picosynth

import "math"

type Signal int32

const (
	SignalBits = 32
	MinSignal  = math.MinInt32
	MaxSignal  = math.MaxInt32
)
