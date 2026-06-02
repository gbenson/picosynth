// Package dbuf coordinates double-buffering.
package dbuf

type Worker struct {
	Name string
	InC  <-chan []uint16
	OutC chan<- []uint16
	ErrC chan<- error
	Func func([]uint16) error
}

func (w *Worker) Start() {
	go func() {
		defer close(w.OutC)

		err := ErrWorkerStopped

		defer func() {
			w.ErrC <- WorkerError{w.Name, err}
		}()

		defer func() {
			if r := recover(); r != nil {
				err = RecoveredPanicError{r}
			}
		}()

		for buf := range w.InC {
			if err = w.Func(buf); err != nil {
				break
			}
			w.OutC <- buf
		}
	}()
}
