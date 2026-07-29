//go:build pico

package piolib

import (
	"errors"
	"machine"
	"math/bits"

	"gbenson.net/go/picosynth/internal/pio"
)

// QuadratureDevice is a wrapper around a PIO state machine that
// tracks the relative position of a quadrature device such as an
// incremental rotary encoder.
type QuadratureDevice struct {
	sm    pio.StateMachine
	last  uint32
	pos   int
	shift int
}

// NewQuadratureDevice creates a new quadrature device peripheral
// using the given PIO state machine.  pinA and pinB must be
// sequential GPIO pins.
func NewQuadratureDevice(sm pio.StateMachine, pinA, pinB machine.Pin) (*QuadratureDevice, error) {
	if pinB != pinA+1 {
		return nil, errors.ErrUnsupported // not sequential.
	}

	sm.TryClaim() // SM should be claimed beforehand, we just guarantee it's claimed.
	Pio := sm.PIO()

	// Program positions.
	const (
		origin     = -1
		entryPoint = 0
		again      = 0
		push_data  = 5
	)
	asm := pio.AssemblerV0{}
	var program = [...]uint16{
		//     .wrap_target
		// again:
		asm.In(pio.InSrcPins, 2).Encode(),                // 0: in pins, 2
		asm.Mov(pio.MovDestX, pio.MovSrcISR).Encode(),    // 1: mov x, isr
		asm.Jmp(pio.JmpXNotEqualY, push_data).Encode(),   // 2: jmp x!=y, push_data
		asm.Mov(pio.MovDestISR, pio.MovSrcNull).Encode(), // 3: mov isr, null
		asm.Jmp(pio.JmpAlways, again).Encode(),           // 4: jmp again
		// push_data:
		asm.Push(false, true).Encode(),              // 5: push
		asm.Mov(pio.MovDestY, pio.MovSrcX).Encode(), // 6: mov y, x
		//     .wrap
	}

	offset, err := Pio.AddProgram(program[:], origin)
	if err != nil {
		return nil, err
	}
	cfg := asm.DefaultStateMachineConfig(offset, program[:])

	// Configure pins
	pinCfg := machine.PinConfig{Mode: Pio.PinMode()}
	pinA.Configure(pinCfg)
	pinB.Configure(pinCfg)

	cfg.SetInPins(pinA, 2)
	cfg.SetFIFOJoin(pio.FifoJoinRx) // merge the FIFOs (we aren't doing any TX...)

	sm.Init(offset, cfg)

	sm.SetPindirsConsecutive(pinA, 2, false)
	sm.Exec(asm.Set(pio.SetDestY, 31).Encode()) // set y, 31
	sm.Jmp(pio.JmpAlways, offset+entryPoint)

	d := &QuadratureDevice{sm: sm}
	d.SetUpdateFrequency(20000) // arbitrary (1k wasn't enough, 10k probably was...)
	d.SetDivisor(4)

	// This enables the state machine. Good practice to not require users to do this
	// since they may be confused why nothing is happening.
	d.Enable(true)

	return d, nil
}

// Enable enables or disables the quadrature device peripheral.
func (d *QuadratureDevice) Enable(enabled bool) {
	d.sm.SetEnabled(enabled)
}

// SetUpdateFrequency sets the update frequency of the quadrature device peripheral.
func (d *QuadratureDevice) SetUpdateFrequency(freq uint32) error {
	whole, frac, err := pio.ClkDivFromFrequency(freq, machine.CPUFrequency())
	if err != nil {
		return err
	}
	d.sm.SetClkDiv(whole, frac)
	return nil
}

// Divisor returns the divisor of the quadrature signal.
func (d *QuadratureDevice) Divisor() int {
	return 1 << d.shift
}

// SetDivisor sets the divisor of the quadrature signal.
func (d *QuadratureDevice) SetDivisor(v int) {
	if v < 1 {
		panic("invalid divisor")
	}

	s := bits.Len(uint(v)) - 1
	if (1 << s) != v {
		panic("invalid divisor")
	}

	d.shift = s
}

// Position returns the accumulated position of the quadrature device.
func (d *QuadratureDevice) Position() int {
	d.pos += d.read()
	return d.pos >> d.shift
}

// SetPosition sets the accumulated position of the quadrature device.
func (d *QuadratureDevice) SetPosition(p int) {
	d.pos = p << d.shift
}

// read returns accumulated movement since its last call.
func (d *QuadratureDevice) read() int {
	if d.sm.IsRxFIFOEmpty() {
		return 0 // no movement
	}

	this := d.sm.RxGet() >> 30
	transition := (d.last << 2) | this
	d.last = this
	return transitions[transition]
}

// clockwise (positive) is 3,2,0,1; anticlockwise is 3,1,0,2.
var transitions = []int{
	0, 1, -1, 0,
	-1, 0, 0, 1,
	1, 0, 0, -1,
	0, -1, 1, 0,
}
