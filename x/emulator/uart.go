package emulator

// A UART is a UART peripheral.
type UART struct {
	Number int
}

// UARTConfig holds configuration parameters for a UART peripheral.
type UARTConfig struct {
	BaudRate uint32
	TX       Pin
	RX       Pin
}

// UARTParity is the parity setting to be used for UART communication.
type UARTParity uint8

// UARTConfigurer wraps the method that implements [UART.Configure].
type UARTConfigurer interface {
	// ConfigureUART configures a UART peripheral.
	ConfigureUART(uart UART, config UARTConfig) error
}

// UARTFormatSetter wraps the method that implements [UART.SetFormat].
type UARTFormatSetter interface {
	// SetFormat configures the framing for a UART peripheral.
	SetUARTFormat(uart UART, databits, stopbits uint8, parity UARTParity) error
}

// UARTBufferSizer wraps the method that implements [UART.Buffered].
type UARTBufferSizer interface {
	// UARTBuffered returns the number of bytes currently stored
	// in the RX buffer.
	UARTBuffered(uart UART) int
}

// UARTByteReader wraps the method that implements [UART.ReadByte].
type UARTByteReader interface {
	// UARTReadByte reads a single byte from the RX buffer.
	// Returns an error if there is no data in the buffer.
	UARTReadByte(uart UART) (byte, error)
}

// Specify whether [UART.Configure] will return an error unless emulated.
var MustConfigureUART = false

// Configure configures a UART peripheral.
func (uart UART) Configure(config UARTConfig) error {
	e, ok := emulator.(UARTConfigurer)
	switch {
	case ok:
		return e.ConfigureUART(uart, config)
	case MustConfigureUART:
		return NotEmulatedError("UART.Configure")
	default:
		return nil
	}
}

// Specify whether [UART.SetFormat] will return an error unless emulated.
var MustSetUARTFormat = false

// SetFormat configures the framing for a UART peripheral.
func (uart UART) SetFormat(databits, stopbits uint8, parity UARTParity) error {
	e, ok := emulator.(UARTFormatSetter)
	switch {
	case ok:
		return e.SetUARTFormat(uart, databits, stopbits, parity)
	case MustConfigureUART:
		return NotEmulatedError("UART.SetFormat")
	default:
		return nil
	}
}

// Buffered returns the number of bytes currently stored in the RX buffer.
func (uart UART) Buffered() int {
	if e, ok := emulator.(UARTBufferSizer); !ok {
		panic(NotEmulatedError("UART.Buffered"))
	} else {
		return e.UARTBuffered(uart)
	}
}

// ReadByte reads a single byte from the RX buffer. If there is no
// data in the buffer, returns an error.
func (uart UART) ReadByte() (byte, error) {
	if e, ok := emulator.(UARTByteReader); !ok {
		return 0, NotEmulatedError("UART.ReadByte")
	} else {
		return e.UARTReadByte(uart)
	}
}
