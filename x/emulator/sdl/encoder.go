package sdl

import "sync/atomic"

type Encoder struct {
	position atomic.Int64
}

var encoder = &Encoder{}

func OpenEncoder(n int) (*Encoder, error) {
	if n != 0 {
		panic("invalid encoder")
	}
	return encoder, ensureStarted()
}

func (e *Encoder) Position() int {
	return int(e.position.Load())
}
