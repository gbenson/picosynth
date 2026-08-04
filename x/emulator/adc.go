package emulator

// An ADC is a [Pin] that may be configured as an analog input.
type ADC struct {
	Pin Pin
}

// ADCConfig holds configuration parameters for an ADC.
type ADCConfig struct {
}

// ADCInitializer wraps the method that implements [InitADC].
type ADCInitializer interface {
	// InitADC resets the ADC peripheral.
	InitADC()
}

// ADCConfigurer wraps the method that implements [ADC.Configure].
type ADCConfigurer interface {
	// ConfigureADC sets an ADC pin to analog input mode.
	ConfigureADC(a ADC, config ADCConfig) error
}

// ADCGetter wraps the method that implements [ADC.Get].
type ADCGetter interface {
	// GetADC returns a one-shot ADC sample reading.
	GetADC(a ADC) uint16
}

// Specify whether [InitADC] will panic unless emulated.
var MustInitADC = false

// InitADC resets the ADC peripheral.
func InitADC() {
	e, ok := emulator.(ADCInitializer)
	switch {
	case ok:
		e.InitADC()
	case MustInitADC:
		panic(NotEmulatedError("InitADC"))
	}
}

// Specify whether [ADC.Configure] will return an error unless emulated.
var MustConfigureADC = false

// Configure sets an ADC pin to analog input mode.
func (a ADC) Configure(config ADCConfig) error {
	e, ok := emulator.(ADCConfigurer)
	switch {
	case ok:
		return e.ConfigureADC(a, config)
	case MustConfigureADC:
		return NotEmulatedError("ADC.Configure")
	default:
		return nil
	}
}

// Get returns a one-shot ADC sample reading.
func (a ADC) Get() uint16 {
	if e, ok := emulator.(ADCGetter); !ok {
		panic(NotEmulatedError("ADC.Get"))
	} else {
		return e.GetADC(a)
	}
}
