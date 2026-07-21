package emulator

import "gbenson.net/go/picosynth/internal/hw"

type KeySwitchMatrix struct {
	outputs []hw.Pin
	inputs  []hw.Pin

	energized []bool
}

func NewKeySwitchMatrix() *KeySwitchMatrix {
	m := &KeySwitchMatrix{}

	numOutputs, numInputs := keyboard.matrixSize()

	m.outputs = make([]hw.Pin, numOutputs)
	m.inputs = make([]hw.Pin, numInputs)

	for i := range numOutputs {
		m.outputs[i] = &outputPin{m, i}
	}
	for j := range numInputs {
		m.inputs[j] = &inputPin{m, j}
	}

	m.energized = make([]bool, numOutputs)

	return m
}

// Outputs implements [hw.KeySwitchMatrix].
func (m *KeySwitchMatrix) Outputs() []hw.Pin {
	return m.outputs
}

// Inputs implements [hw.KeySwitchMatrix].
func (m *KeySwitchMatrix) Inputs() []hw.Pin {
	return m.inputs
}

type pin struct {
	ksm *KeySwitchMatrix
	num int
}

type outputPin pin
type inputPin pin

func (p *outputPin) Set(level bool) {
	keyboard.mu.Lock()
	defer keyboard.mu.Unlock()

	p.ksm.energized[p.num] = level
}

func (p *inputPin) Get() bool {
	keyboard.mu.Lock()
	defer keyboard.mu.Unlock()

	for i, held := range keyboard.matrix[p.num] {
		if held && p.ksm.energized[i] {
			return true
		}
	}

	return false
}

func (p *outputPin) Get() bool {
	panic("not an input")
}

func (p *inputPin) Set(level bool) {
	panic("not an output")
}
