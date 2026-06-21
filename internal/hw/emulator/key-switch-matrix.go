package emulator

import "gbenson.net/go/picosynth/internal/hw"

type KeySwitchMatrix struct {
	outputs []hw.OutputPin
	inputs  []hw.InputPin

	energized []bool
}

func OpenKeySwitchMatrix() (*KeySwitchMatrix, error) {
	m := &KeySwitchMatrix{}

	numOutputs, numInputs := keyboard.matrixSize()

	m.outputs = make([]hw.OutputPin, numOutputs)
	m.inputs = make([]hw.InputPin, numInputs)

	for i := range numOutputs {
		m.outputs[i] = &outputPin{m, i}
	}
	for j := range numInputs {
		m.inputs[j] = &inputPin{m, j}
	}

	m.energized = make([]bool, numOutputs)

	return m, ensureStarted()
}

// Outputs implements [hw.KeySwitchMatrix].
func (m *KeySwitchMatrix) Outputs() []hw.OutputPin {
	return m.outputs
}

// Inputs implements [hw.KeySwitchMatrix].
func (m *KeySwitchMatrix) Inputs() []hw.InputPin {
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
