package emulator

import (
	"time"

	"gbenson.net/go/picosynth/internal/hw"
)

type KeySwitchMatrix struct {
	outStates [5]bool
	inStates  [4]bool

	outputs []hw.OutputPin
	inputs  []hw.InputPin

	arp arpeggiator
}

func OpenKeySwitchMatrix() (*KeySwitchMatrix, error) {
	kb := &KeySwitchMatrix{
		arp: arpeggiator{
			Notes: []uint8{48, 52, 55, 59, 60, 59, 55, 52},
			Tempo: 180 * time.Millisecond,
		},
	}

	kb.outputs = make([]hw.OutputPin, len(kb.outStates))
	kb.inputs = make([]hw.InputPin, len(kb.inStates))

	for i := range kb.outStates {
		kb.outputs[i] = &outputPin{kb, i}
	}
	for j := range kb.inStates {
		kb.inputs[j] = &inputPin{kb, j}
	}
	return kb, nil
}

// Outputs implements [hw.KeySwitchMatrix].
func (kb *KeySwitchMatrix) Outputs() []hw.OutputPin {
	return kb.outputs
}

// Inputs implements [hw.KeySwitchMatrix].
func (kb *KeySwitchMatrix) Inputs() []hw.InputPin {
	return kb.inputs
}

type pin struct {
	kb  *KeySwitchMatrix
	num int
}

type outputPin pin
type inputPin pin

func (p *outputPin) Set(level bool) {
	p.kb.outStates[p.num] = level
	p.kb.update()
}

func (p *inputPin) Get() bool {
	return p.kb.inStates[p.num]
}

func (kb *KeySwitchMatrix) update() {
	kb.arp.Step()
	midinote := kb.arp.Note()      // 48..60
	scancode := int(midinote) - 41 //  7..19

	// clear input pins
	for j := range kb.inStates {
		kb.inStates[j] = false
	}

	output := scancode / len(kb.inStates)
	if !kb.outStates[output] {
		return // no input pins high
	}

	kb.inStates[scancode%len(kb.inStates)] = true
}

type arpeggiator struct {
	Notes       []uint8
	Tempo       time.Duration
	note        int
	lastStepped time.Time
}

func (arp *arpeggiator) Step() {
	now := time.Now()
	if now.Sub(arp.lastStepped) < arp.Tempo {
		return
	} else if !arp.lastStepped.IsZero() {
		arp.note = (arp.note + 1) % len(arp.Notes)
	}
	arp.lastStepped = now
}

func (arp *arpeggiator) Note() uint8 {
	return arp.Notes[arp.note]
}
