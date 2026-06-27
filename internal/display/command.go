package display

type Command string

const (
	ClearCommand     Command = "\x0c" // clear the buffer
	SleepCommand     Command = "\x04" // power down the display
	SyncCommand      Command = "\x0a" // push the buffer to the display
	KeepAliveCommand Command = "\x06" // wake display/inhibit screensaver
)

// NewTextCommand creates a command to display the given text,
// expanded to fill the entire screen.
func NewTextCommand(s string) Command {
	// This command is the expected use-case for the display, for
	// real-time feedback while playing.  It should operate without
	// allocation.
	return Command(s)
}

// NewTextIfSerialCommand creates a command to display the given text,
// expanded to fill the entire screen, but only [Display.serial]
// matches the supplied value at the time the command is received.
func NewTextIfSerialCommand(n int32, s string) Command {
	return Command("\x1B" + i2s(n) + s)
}

// NewTextAtCommand creates a command to display the given text with
// its top left corner at x, y, roughly h pixels high.
func NewTextAtCommand(x, y, h int32, s string) Command {
	return Command("\x02" + i2s(x) + i2s(y) + i2s(h) + s)
}

// NewLineCommand creates a command to draw a line from x1, y1 to x2, y2.
func NewLineCommand(x1, y1, x2, y2 int32) Command {
	return Command("\x01" + i2s(x1) + i2s(y1) + i2s(x2) + i2s(y2))
}
