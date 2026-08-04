package emulator

type NotEmulatedError string

func (e NotEmulatedError) Error() string {
	return string(e) + ": not emulated"
}
