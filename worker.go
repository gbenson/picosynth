package picosynth

import (
	"io"
	"sync/atomic"

	"gbenson.net/go/picosynth/internal/hw"
)

type worker interface {
	Name() string
	Run() func() error
}

type workerManager struct {
	count atomic.Int32
	errC  chan error
}

// newWorkerManager returns a manager that can start up to n workers.
func newWorkerManager(n int) *workerManager {
	wm := &workerManager{errC: make(chan error, n*2)}
	wm.count.Store(int32(n))
	return wm
}

func (wm *workerManager) Start(w worker) {
	if wm.count.Add(-1) < 0 {
		panic("too many workers")
	}
	go func() {
		if c, ok := w.(io.Closer); ok {
			defer wm.invoke(w, c.Close, nil)
		}

		wm.invoke(w, w.Run(), ErrWorkerStopped)
	}()
}

// invoke calls f on behalf of w.
func (wm *workerManager) invoke(w worker, f func() error, defaultError error) {
	var err error

	defer func() {
		if err == nil {
			err = defaultError
		}
		if err != nil {
			wm.errC <- WorkerError{w.Name(), err}
		}
	}()

	if hw.IsBareMetal {
		defer func() {
			if r := recover(); r != nil {
				err = RecoveredPanicError{r}
			}
		}()
	}

	err = f()
}

// Wait returns the first error returned by any worker.
func (wm *workerManager) Wait() error {
	return <-wm.errC
}
