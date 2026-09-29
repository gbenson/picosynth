//go:build rp2350

package ksm

import (
	"device/rp"
	"runtime/volatile"
	"unsafe"

	"gbenson.net/go/picosynth/internal/hw/machine"
)

// copy-pasted from tinygo/src/machine/machine_rp2_2350.go
const _NUMBANK0_GPIOS = 48

// copy-pasted from tinygo/src/machine/machine_rp2_gpio.go
type padsBank0Type struct {
	voltageSelect volatile.Register32
	io            [_NUMBANK0_GPIOS]volatile.Register32
}

var padsBank0 = (*padsBank0Type)(unsafe.Pointer(rp.PADS_BANK0))

func padCtrl(p machine.Pin) *volatile.Register32 {
	return &padsBank0.io[p]
}

// end copy-pasted code

const inputEnableMask = rp.PADS_BANK0_GPIO0_IE_Msk

func enableInputBuffer(p machine.Pin) {
	padCtrl(p).SetBits(inputEnableMask)
}

func disableInputBuffer(p machine.Pin) {
	padCtrl(p).ClearBits(inputEnableMask)
}

// configurePin configures a pin.
func configurePin(p machine.Pin, config machine.PinConfig) {
	p.Configure(config)

	// RP2350 datasheet, Errata 9 workaround: "If pad pull-down behaviour
	// is required, clear the pad input enable in GPIO0.IE to ensure that
	// the pad pull-down resistor pulls the pad signal low."
	if config.Mode == machine.PinInputPulldown {
		disableInputBuffer(p)
	}
}

// Read reads the pin value.
func Read(p machine.Pin) bool {
	// RP2350 datasheet, Errata 9 workaround: "To read the state of a
	// pad pulled-down GPIO from software, enable the input buffer by
	// setting GPIO0.IE immediately before reading, and then re-disable
	// immediately afterwards."
	enableInputBuffer(p)
	defer disableInputBuffer(p)

	return p.Get()
}
