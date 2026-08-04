package emulator

// An I2C is an I2C interface peripheral.
type I2C int

// I2CConfig holds configuration parameters for an I2C peripheral.
type I2CConfig struct {
	SDA, SCL Pin
}

// I2CConfigurer wraps the method that implements [I2C.Configure].
type I2CConfigurer interface {
	// ConfigureI2C initializes an I2C peripheral and configures its pins.
	ConfigureI2C(i2c I2C, config I2CConfig) error
}

// I2CTxer wraps the method that implements [I2C.Tx].
type I2CTxer interface {
	// I2CTx performs a write and then a read transfer, placing
	// the result in r.  Passing a nil value for w or r skips the
	// corresponding transfer.
	I2CTx(i2c I2C, addr uint16, w, r []byte) error
}

// Specify whether [I2C.Configure] will return an error unless emulated.
var MustConfigureI2C = false

// Configure initializes an I2C peripheral and configures its pins.
func (i2c I2C) Configure(config I2CConfig) error {
	e, ok := emulator.(I2CConfigurer)
	switch {
	case ok:
		return e.ConfigureI2C(i2c, config)
	case MustConfigureI2C:
		return NotEmulatedError("I2C.Configure")
	default:
		return nil
	}
}

// Tx performs a write and then a read transfer, placing the result
// in r.  Passing a nil value for w or r skips the corresponding
// transfer.
func (i2c I2C) Tx(addr uint16, w, r []byte) error {
	if e, ok := emulator.(I2CTxer); !ok {
		return NotEmulatedError("I2C.Tx")
	} else {
		return e.I2CTx(i2c, addr, w, r)
	}
}
