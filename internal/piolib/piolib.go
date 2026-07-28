package piolib

import "runtime"

func gosched() {
	runtime.Gosched()
}
