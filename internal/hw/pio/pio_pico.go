//go:build pico

package pio

import "github.com/tinygo-org/pio/rp2-pio"

type StateMachine = pio.StateMachine

var (
	PIO0 = pio.PIO0
	PIO1 = pio.PIO1
)
