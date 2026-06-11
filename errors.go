package picosynth

import (
	"errors"

	"gbenson.net/go/picosynth/internal/fmt"
)

var ErrWorkerStopped = errors.New("worker stopped")

type WorkerError struct {
	worker string
	err    error
}

func (e WorkerError) Error() string {
	return e.worker + ": " + e.err.Error()
}

func (e WorkerError) Unwrap() error {
	return e.err
}

type RecoveredPanicError struct {
	v any
}

func (e RecoveredPanicError) Error() string {
	return fmt.Sprint("panic: ", e.v)
}
