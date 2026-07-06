package display

type Opcode int32

const (
	// Zero-operand instructions have negative opcodes.
	// For these instructions the value transmitted over
	// the channel is the opcode itself.

	CmdClear Opcode = -1 - iota // clear the buffer
	CmdSleep                    // power down the display
	CmdSync                     // push the buffer to the display
	CmdWake                     // wake display/inhibit screensaver

	// Instructions with operands have positive opcodes.
	// For these instructions the value transmitted over
	// the channel is an index into the instruction buffer,
	// with the opcode stored in the referenced entry.

	CmdBox  Opcode = iota + 1 // draw a filled box
	CmdText                   // render arbitrary text
)

type Instruction struct {
	Opcode     Opcode // 4 bytes
	IfSerial   int32  // 4 bytes
	X, Y, W, H int32  // 16 bytes
	Int32      string // 4 bytes
	String     string // 4 bytes
}

// A command is a zero-argument opcode, if negative, or
// an index into a display's instruction buffer otherwise.
type command int32
