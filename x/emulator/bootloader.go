package emulator

// EnterBootloader resets the system into the USB bootloader.
func EnterBootloader() {
	panic(NotEmulatedError("EnterBootloader"))
}
