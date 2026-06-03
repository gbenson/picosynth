//go:build tinygo

package main

import (
	"machine"
	"time"
)

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
