package picosynth

import "runtime"

func gosched() {
	runtime.Gosched()
}
