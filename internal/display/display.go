package display

import (
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/ssd1306"
)

const (
	Width   = 128
	Height  = 32
	Address = ssd1306.Address_128_32

	bufsiz = Width * Height / 8
)

type Display struct {
	bus    drivers.I2C
	device *ssd1306.Device
	buffer [bufsiz]byte
}

// Name implements [worker].
func (d *Display) Name() string {
	return "display"
}

// Run implements [worker].
func (d *Display) Run() func() error {
	return d.run
}

func (d *Display) run() error {
	if d.bus != nil {
		panic("already started")
	}

	if bus, err := openBus(); err != nil {
		return err
	} else {
		d.bus = bus
	}

	d.device = ssd1306.NewI2C(d.bus)
	d.device.Configure(ssd1306.Config{
		Width:   Width,
		Height:  Height,
		Address: Address,
	})

	// XXX >>>
	if err := d.Show("Hello!"); err != nil {
		return err
	}
	time.Sleep(time.Second << 32)
	// <<< XXX
	return nil
}

// Show displays the given text.
func (d *Display) Show(text string) error {
	buf := d.buffer[:]
	for i := range bufsiz {
		buf[i] = byte(i+1) ^ byte((i+1)>>8)
	}
	if err := d.device.SetBuffer(buf); err != nil {
		return err
	}
	return d.device.Display()
}
