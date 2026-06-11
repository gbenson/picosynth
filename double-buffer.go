package picosynth

type doubleBuffer[T any] struct {
	Filler, Player worker
}

type dbufWorker[T any] struct {
	name string
	inC  <-chan []T
	outC chan<- []T
	Func func([]T) error
}

func newDoubleBuffer[T any](size int, fill, play func([]T) error) *doubleBuffer[T] {
	buffers := make([]T, size*2)

	fillMe := make(chan []T, 2)
	playMe := make(chan []T)

	fillMe <- buffers[:size]
	fillMe <- buffers[size:]

	return &doubleBuffer[T]{
		Filler: &dbufWorker[T]{
			name: "filler",
			inC:  fillMe,
			outC: playMe,
			Func: fill,
		},
		Player: &dbufWorker[T]{
			name: "player",
			inC:  playMe,
			outC: fillMe,
			Func: play,
		},
	}
}

// Name implements [worker].
func (w *dbufWorker[T]) Name() string {
	return w.name
}

// Run implements [worker].
func (w *dbufWorker[T]) Run() func() error {
	return w.run
}

func (w *dbufWorker[T]) run() (err error) {
	for buf := range w.inC {
		if err = w.Func(buf); err != nil {
			break
		}
		w.outC <- buf
	}
	return
}

// Close implements [io.Closer].
func (w *dbufWorker[T]) Close() error {
	close(w.outC)
	return nil
}
