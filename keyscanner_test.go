package picosynth

import (
	"errors"
	"testing"

	"gotest.tools/v3/assert"

	"gbenson.net/go/picosynth/internal/hw/machine"
	"gbenson.net/go/picosynth/internal/keyboard"
)

func TestCasioScanner(t *testing.T) {
	const (
		SET uint8 = 254 - iota
		CLR
		GET
	)

	want := make([]uint8, 0, 512)
	got := make([]uint8, 0, 512)

	ksm := keyboard.Matrix()

	for _, ko := range ksm.Outputs {
		want = append(want, CLR, uint8(ko))
	}
	for _ = range 2 {
		for _, ko := range ksm.Outputs {
			want = append(want, SET, uint8(ko))

			for _, ki := range ksm.Inputs {
				want = append(want, GET, uint8(ki))
			}

			want = append(want, CLR, uint8(ko))
		}
	}
	assert.Check(t, len(want) > 256 && len(want) < 512)

	ks := &KeyScanner{}
	em := &ksTestEmulator{
		setPin: func(p machine.Pin, v bool) {
			for e := ks.Poll(); e != NoEvent; e = ks.Poll() {
				t.Log("unexpected event", e)
				t.Fail()
			}

			if len(got) >= len(want) {
				StopTest()
			}

			if v {
				got = append(got, SET, uint8(p))
			} else {
				got = append(got, CLR, uint8(p))
			}
		},
		getPin: func(p machine.Pin) bool {
			got = append(got, GET, uint8(p))
			return false
		},
	}

	defer func() {
		rv := recover()
		if err, _ := rv.(error); !errors.Is(err, ErrTestComplete) {
			t.Log(rv)
			t.FailNow()
		}

		assert.Equal(t, em.inputs, uint(0xf0f0))
		assert.Equal(t, em.outputs, uint(0x7f0000))

		if string(got[:len(want)]) != string(want) {
			t.Log("want:", want)
			t.Log("got: ", got)
			t.Fail()
		}
	}()

	WithTestEmulator(t, em)
	assert.NilError(t, ks.run())
}

type ksTestEmulator struct {
	inputs, outputs uint

	setPin func(machine.Pin, bool)
	getPin func(machine.Pin) bool
}

func (e *ksTestEmulator) ConfigurePin(p machine.Pin, config machine.PinConfig) {
	mask := uint(1) << p
	if (e.inputs|e.outputs)&mask != 0 {
		panic("pin already configured")
	}

	switch config.Mode {
	case machine.PinInputPulldown:
		e.inputs |= mask
	case machine.PinOutput:
		e.outputs |= mask
	default:
		panic("invalid mode")
	}
}

func (e *ksTestEmulator) SetPin(p machine.Pin, v bool) {
	mask := uint(1) << p
	if e.outputs&mask == 0 {
		panic("pin not configured for output")
	}

	e.setPin(p, v)
}

func (e *ksTestEmulator) GetPin(p machine.Pin) bool {
	mask := uint(1) << p
	if e.inputs&mask == 0 {
		panic("pin not configured for input")
	}
	return e.getPin(p)
}
