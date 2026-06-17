//go:build pico

package display

import "machine"

func openBus() (*machine.I2C, error) {
	i2c := machine.I2C1
	err := i2c.Configure(machine.I2CConfig{
		SDA: machine.GPIO2,
		SCL: machine.GPIO3,
	})
	if err != nil {
		return nil, err
	}
	return i2c, nil
}
