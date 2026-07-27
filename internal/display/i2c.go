package display

import (
	"runtime"

	"tinygo.org/x/drivers"
)

type I2C struct {
	bus drivers.I2C
}

// Tx implements [drivers.I2C].
func (d *I2C) Tx(addr uint16, w, r []byte) error {
	const dataMode = 0x40
	if r != nil || len(w) != bufsiz+1 || w[0] != dataMode {
		return d.bus.Tx(addr, w, r)
	}

	const chunkShift = 6
	const chunkSize = 1 << chunkShift
	for i := 0; i < bufsiz; i += chunkSize {
		runtime.Gosched()

		saved := w[i]
		w[i] = dataMode
		err := d.bus.Tx(addr, w[i:i+chunkSize+1], nil)
		w[i] = saved
		if err != nil {
			return err
		}
	}

	return nil
}
