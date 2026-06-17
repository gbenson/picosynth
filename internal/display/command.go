package display

type Command string

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
	return Command("\x1B" + string(rune(n)) + s)
}

// future commands:
//  - Clear() - wipe the buffer
//  - TextAt(x, y int, s string) - render unexpanded text
//      (maybe with y limited to page boundaries, idk)
//  - Sync()/Update() - push the buffer to the display.
//      (if added, update Text() docstring to detail that
//      Text() includes a Sync(), that's only there for if
//      you're constructing a menu/QR code/whatever.
//
// these can be packed into type Command as strings whose
// first character is some special character ('\x1B'?) that
// [Display.do] can detect, or maybe have a [Command.Decode]
// method that takes any command apart (including Text)
