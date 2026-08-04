package sdl

import "sync/atomic"

type Encoder struct {
	position atomic.Int64
}

var encoder = &Encoder{}

func (e *Encoder) Position() int {
	return int(e.position.Load())
}
