package piolib

import (
	"errors"
	"runtime"
)

var errBusy = errors.New("piolib:busy")

func gosched() {
	runtime.Gosched()
}
