//go:build baremetal

package main

import (
	"machine"
	"time"

	"gbenson.net/go/picosynth"
)

func main() {
	defer func() { fatal(recover()) }()

	if err := picosynth.Run(); err != nil {
		fatal(err)
	}

	fatal("unexpected nil error")
}

// fatal prints a message and slow-flashes the LED forever.
func fatal(msg any) {
	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	for {
		println(msg)

		led.High()
		time.Sleep(time.Second / 2)

		led.Low()
		time.Sleep(time.Second / 2)
	}
}
