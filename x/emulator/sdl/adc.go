package sdl

import (
	"math"
	"time"
)

type ADC int

func OpenADC(n int) (ADC, error) {
	return ADC(n), ensureStarted()
}

func (n ADC) Get() uint16 {
	mask := int64((1 << (8 + n*2)) - 1)
	mt := time.Now().UnixMilli() & mask
	phase := float64(mt) / float64(mask+1) * math.Pi
	v := int(math.Round(1024 * (1 + math.Sin(phase))))
	return uint16(max(0, min(2048, v))) << 4
}
