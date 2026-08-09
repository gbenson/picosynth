package picosynth

import (
	"time"

	"gbenson.net/go/picosynth/internal/hw/machine"
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

	if err := keyboard.Configure(); err != nil {
		return err
	}

	ksm := keyboard.Matrix()
	rows := ksm.Outputs
	cols := ksm.Inputs

	for _, rp := range rows {
		ks.setPin(rp, false)
	}

	lastState := make([]bool, len(rows)*len(cols))
	settleTime := maxLatency / time.Duration(len(rows))

	for {
		var sc Scancode

		for _, rp := range rows {
			ks.setPin(rp, true)
			time.Sleep(settleTime)

			for _, cp := range cols {
				state := cp.Get()
				if state != lastState[sc] {
					events <- keyEvent(sc, state)
					lastState[sc] = state
				}
				sc++
			}

			ks.setPin(rp, false)
		}
	}
}

func (ks *KeyScanner) setPin(p machine.Pin, v bool) {
	if p == machine.NoPin {
		return
	}
	p.Set(v)
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
