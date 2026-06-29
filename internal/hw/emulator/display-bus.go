package emulator

import (
	"errors"
	"fmt"

	"tinygo.org/x/drivers/ssd1306"
)

type DisplayBus struct {
	cmds []byte
}

func OpenDisplayBus() (*DisplayBus, error) {
	return &DisplayBus{}, nil
}

func (d *DisplayBus) Tx(addr uint16, w, r []byte) error {
	if len(r) != 0 {
		return fmt.Errorf("%w: r=%v", errors.ErrUnsupported, r)
	}

	var err error
	switch w[0] {
	case 0x00:
		err = d.onCommand(w[1:])
	case 0x40:
		err = d.onData(w[1:])
	default:
		err = errors.ErrUnsupported
	}

	if errors.Is(err, errors.ErrUnsupported) {
		err = fmt.Errorf("%w: w=%v", err, w)
	}

	return err
}

func (d *DisplayBus) onCommand(buf []byte) error {
	if len(buf) != 1 {
		return errors.ErrUnsupported
	}
	cmd := buf[0]
	switch cmd {
	case ssd1306.DISPLAYOFF:
		display.Sleep()
	case ssd1306.DISPLAYON:
		display.Wake()
	}
	d.cmds = append(d.cmds, cmd)
	return nil
}

func (d *DisplayBus) onData(buf []byte) error {
	// last commands should be the six byte reset sequence
	// from tinygo.org/x/drivers/ssd1306/Device.Display.
	if len(d.cmds) < 6 {
		return errors.ErrUnsupported
	}
	setup := d.cmds[len(d.cmds)-6:]
	switch {
	case setup[0] != ssd1306.COLUMNADDR:
	case setup[1] != 0:
	case setup[3] != ssd1306.PAGEADDR:
	case setup[4] != 0:
	default:
		display.SetBuffer(int32(setup[2])+1, (int32(setup[5])+1)*8, buf)
		return ensureStarted()
	}

	return errors.ErrUnsupported
}
