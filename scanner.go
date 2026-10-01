package picosynth

import (
	"time"

	"gbenson.net/go/picosynth/ksm"
)

// Scanner scans the keyboard matrix.
type Scanner struct {
	events <-chan KeyEvent
}

// Name implements [worker].
func (ks *Scanner) Name() string {
	return "scanner"
}

// Run implements [worker].
func (ks *Scanner) Run() func() error {
	return ks.run
}

func (ks *Scanner) run() error {
	if ks.events != nil {
		panic("already started")
	}

	events := make(chan KeyEvent)
	defer close(events)
	ks.events = events

	if err := ksm.Configure(); err != nil {
		return err
	}

	rows := ksm.Outputs
	cols := ksm.Inputs

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
				state := ksm.Read(cp)
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
func (ks *Scanner) Poll() KeyEvent {
	select {
	case event := <-ks.events:
		return event
	default:
		return NoEvent
	}
}
