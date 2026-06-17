package emulator

type DisplayBus struct {
}

func OpenDisplayBus() (*DisplayBus, error) {
	return &DisplayBus{}, nil
}

func (d *DisplayBus) Tx(addr uint16, w, r []byte) error {
	panic("not implemented")
}
