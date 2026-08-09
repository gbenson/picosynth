package uart

import (
	"errors"
	"time"

	"gbenson.net/go/picosynth/internal/hw/machine"
)

type UART struct {
	uart     *machine.UART
	waitTime uint32
}

// Open returns a UART configured with the given baudrate and pins.
func Open(tx, rx machine.Pin, baud uint32) (*UART, error) {
	config := machine.UARTConfig{BaudRate: baud, TX: tx, RX: rx}

	// How long to wait if we want to read but nothing's there?
	bitTime := uint32(time.Second) / baud // 32 μs/bit for 31240 baud MIDI
	waitTime := bitTime + bitTime/8       // 36 μs/bit for 31240 baud MIDI

	var errs []error
	for _, d := range []*machine.UART{machine.UART0, machine.UART1} {
		err := d.Configure(config)
		if err == nil {
			return &UART{d, waitTime}, nil
		}
		errs = append(errs, err)
	}

	return nil, errors.Join(errs...)
}

// SetFormat configures the framing for the UART.
func (u *UART) SetFormat(
	databits, stopbits uint8,
	parity machine.UARTParity,
) error {
	return u.uart.SetFormat(databits, stopbits, parity)
}

// ReadByte reads and returns the next byte from the UART.
func (u *UART) ReadByte() (byte, error) {
	for u.uart.Buffered() < 1 {
		time.Sleep(time.Duration(u.waitTime))
	}
	return u.uart.ReadByte()
}
