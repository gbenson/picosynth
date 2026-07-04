package adc

import "gbenson.net/go/picosynth/internal/hw"

type ADC = hw.ADC

// Open opens the specified ADC.
func Open(n int) (ADC, error) {
	return open(n)
}
