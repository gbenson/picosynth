package emulator

// An Encoder is an incremental encoder connected to a pair of pins.
type Encoder struct {
	PinA, PinB Pin
}

// EncoderConfig holds configuration parameters for an Encoder.
type EncoderConfig struct {
	Precision int
}

// NewQuadratureViaInterrupt creates and initializes a new Encoder.
func NewQuadratureViaInterrupt(pinA, pinB Pin) *Encoder {
	return &Encoder{PinA: pinA, PinB: pinB}
}

// EncoderConfigurer wraps the method that implements [Encoder.Configure].
type EncoderConfigurer interface {
	// ConfigureEncoder configures the encoder's pins.
	ConfigureEncoder(enc Encoder, config EncoderConfig) error
}

// EncoderPositioner wraps the method that implements [Encoder.Position].
type EncoderPositioner interface {
	// EncoderPosition returns the accumulated position of the encoder.
	EncoderPosition(enc Encoder) int
}

// Specify whether [Encoder.Configure] will return an error unless emulated.
var MustConfigureEncoder = false

// Configure configures the encoder's pins.
func (enc Encoder) Configure(config EncoderConfig) error {
	e, ok := emulator.(EncoderConfigurer)
	switch {
	case ok:
		return e.ConfigureEncoder(enc, config)
	case MustConfigureEncoder:
		return NotEmulatedError("Encoder.Configure")
	default:
		return nil
	}
}

// Position returns the  accumulated position of the encoder.
func (enc Encoder) Position() int {
	if e, ok := emulator.(EncoderPositioner); !ok {
		panic(NotEmulatedError("Encoder.Position"))
	} else {
		return e.EncoderPosition(enc)
	}
}
