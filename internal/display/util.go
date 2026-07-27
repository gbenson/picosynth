package display

import "runtime"

func gosched() {
	runtime.Gosched()
}
