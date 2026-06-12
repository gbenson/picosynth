package picosynth

import (
	"time"

	"gbenson.net/go/picosynth/internal/keyboard"
)

type KeyScanner struct {
	events <-chan KeyEvent
}

// Name implements [worker].
func (ks *KeyScanner) Name() string {
	return "keyscanner"
}

// Run implements [worker].
func (ks *KeyScanner) Run() func() error {
	return ks.run
}

func (ks *KeyScanner) run() error {
	if ks.events != nil {
		panic("already started")
	}

	events := make(chan KeyEvent)
	defer close(events)
	ks.events = events

	kb := keyboard.New()

	rows := kb.Rows()
	cols := kb.Columns()

	for _, rp := range rows {
		rp.Set(false)
	}

	lastState := make([]bool, len(rows)*len(cols))
	settleTime := maxLatency / time.Duration(len(rows))

	for {
		var sc Scancode

		for _, rp := range rows {
			rp.Set(true)
			time.Sleep(settleTime)

			for _, cp := range cols {
				state := cp.Get()
				if state != lastState[sc] {
					events <- keyEvent(sc, state)
					lastState[sc] = state
				}
				sc++
			}

			rp.Set(false)
		}
	}
}

// Poll returns the next pending event, or NoEvent if there are none.
func (ks *KeyScanner) Poll() KeyEvent {
	select {
	case event := <-ks.events:
		return event
	default:
		return NoEvent
	}
}
