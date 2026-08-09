package emulator

// A Pin is a single pin on a chip. It can be used directly, as
// a GPIO, or indirectly via peripherals such as ADC, I2C, etc.
type Pin uint8

// NoPin explicitly indicates "not a pin".
const NoPin = Pin(0xff)

// PinMode describes the direction and pull mode of a Pin.
type PinMode uint8

// PinConfig holds configuration parameters for a Pin.
type PinConfig struct {
	Mode PinMode
}

// PinConfigurer wraps the method that implements [Pin.Configure].
type PinConfigurer interface {
	// ConfigurePin configures a GPIO pin.
	ConfigurePin(p Pin, config PinConfig)
}

// PinGetter wraps the method that implements [Pin.Get].
type PinGetter interface {
	// GetPin reads the pin value.
	GetPin(p Pin) bool
}

// PinSetter wraps the method that implements [Pin.Set].
type PinSetter interface {
	// SetPin sets the pin value.
	SetPin(p Pin, v bool)
}

// Specify whether [Pin.Configure] will panic unless emulated.
var MustConfigurePin = false

// Configure configures a GPIO pin.
func (p Pin) Configure(config PinConfig) {
	e, ok := emulator.(PinConfigurer)
	switch {
	case ok:
		e.ConfigurePin(p, config)
	case MustConfigurePin:
		panic(NotEmulatedError("Pin.Configure"))
	}
}

// Get reads the pin value.
func (p Pin) Get() bool {
	if e, ok := emulator.(PinGetter); !ok {
		panic(NotEmulatedError("Pin.Get"))
	} else {
		return e.GetPin(p)
	}
}

// Set sets the pin value.
func (p Pin) Set(v bool) {
	if e, ok := emulator.(PinSetter); !ok {
		panic(NotEmulatedError("Pin.Set"))
	} else {
		e.SetPin(p, v)
	}
}
