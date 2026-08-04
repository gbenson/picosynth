package emulator

// An I2S is an I2S peripheral.
type I2S struct {
	StateMachine StateMachine
	Data, BClk   Pin
}

// NewI2S creates a new I2S peripheral using the given state machine.
func NewI2S(sm StateMachine, data, clockAndNext Pin) (*I2S, error) {
	return &I2S{StateMachine: sm, Data: data, BClk: clockAndNext}, nil
}

// I2SSampleFrequencySetter wraps the method that implements [I2S.SetSampleFrequency].
type I2SSampleFrequencySetter interface {
	// I2SSetSampleFrequency sets the sample frequency of an I2S peripheral.
	I2SSetSampleFrequency(i2s *I2S, freq uint32) error
}

// I2SMonoWriter wraps the method that implements [I2S.WriteMono].
type I2SMonoWriter interface {
	// WriteMono writes a mono audio buffer to an I2S peripheral.
	I2SWriteMono(i2s *I2S, b []uint16) (int, error)
}

// SetSampleFrequency sets the sample frequency of the I2S peripheral.
func (i2s *I2S) SetSampleFrequency(freq uint32) error {
	if e, ok := emulator.(I2SSampleFrequencySetter); !ok {
		return NotEmulatedError("I2S.SetSampleFrequency")
	} else {
		return e.I2SSetSampleFrequency(i2s, freq)
	}
}

// WriteMono writes a mono audio buffer to the I2S peripheral.
func (i2s *I2S) WriteMono(b []uint16) (int, error) {
	if e, ok := emulator.(I2SMonoWriter); !ok {
		return 0, NotEmulatedError("I2S.WriteMono")
	} else {
		return e.I2SWriteMono(i2s, b)
	}
}
