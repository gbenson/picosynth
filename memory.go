package picosynth

// A Memory comprises 512 numbered 32-bit registers laid out as
// follows:
//
//	| Register  | Purpose                      |
//	| --------- | ---------------------------- |
//	| 0x00–0x3f | 64 general purpose registers |
//	| 0x40–0x7f | 64 matrix bias registers     |
//	| 0x80–0xbf | 64 feedback registers        |
//	| 0xc0–0xff | 64 matrix output registers   |
//	| 0x100...  | 64×4 matrix connections      |
//
// Or:
//   - 128 parameter registers (0x00..0x7f).  These are intended to
//     have their values set by control inputs, but may be edited
//     in the memory editor in the meantime.  By convention the
//     upper half of this block (registers 0x40..0x7f) are biases
//     for the 64 mod matrix rows; the matrix is initialized with
//     the first cell of each row set as its bias input at 100% level
//     and other cells zeroed such such that e.g. the cutoff input
//     to the filter (the value in the ModulatedFilterCutoff register)
//     will be the value of the FilterCutoff bias input in the
//     absence of other modulations (with the value of the FilterCutoff
//     bias input being updated with the value read from its dedicated
//     hardware knob each timestep.
//   - 64 feedback registers (0x80..0xbf).  These are outputs from
//     synth components for feeding back into mod matrix calculations.
//   - 64 matrix output registers (0xc0..0xff).  These are calculated
//     on demand using the mod matrix, up to once per step, so they're
//     not calculated if they're not used, and they may include values
//     from previous timesteps if e.g. a feedback register is read
//     before the point it would be set in the current timestep.
//   - 64 rows of 4 mod matrix entries each, with each entry being
//     an 8-bit register number in the upper 8 bits and a 24-bit
//     signed Q1.23 fixed point amount to multiply the referenced
//     register's contents with.  If the amount is 0 then the register
//     won't be read.  Note that if you need more than 4 modulations
//     for one output you can chain rows using the feedback registers.
type Memory struct {
	registers [NumRegisters]uint32 // 2048 bytes

	serial  uint
	serials [NumRegisters]uint // also 2048 bytes (on 32-bit)
	// These could be bytes, if memory becomes an issue, or maybe
	// even single bits, but they're ints for speed for the time
	// being. This might be different on non-RP2040 hardware, ymmv.
}

func (m *Memory) Register(n int) Register {
	return Register{m, n}
}

func (m *Memory) Load(n int) uint32 {
	m.maybeStepRegister(n)
	return m.load(n)
}

func (m *Memory) load(n int) uint32 {
	return m.registers[n]
}

func (m *Memory) Store(n int, v uint32) {
	m.registers[n] = v
}

func (m *Memory) Step() {
	m.serial++
}

func (m *Memory) maybeStepRegister(n int) {
	if (n >> 6) != 3 {
		return // not a mod matrix output (n<0xc0 || n>0xff)
	}
	if m.serials[n] == m.serial {
		return // already stepped
	}
	m.serials[n] = m.serial // update now, in case of loops

	srcBase := (n & 0x7f) << 2 // 0xc0 => 0x100, 0xc1 => 0x104, etc

	var result Signal
	for i := range 4 {
		c := matrixCell(m.registers[srcBase+i])
		if c == 0 {
			continue // empty cell
		}
		src := m.Register(c.Source())
		result += c.Scale().Mul(Signal(src.Load()))
	}

	dst := m.Register(n)
	dst.Store(uint32(result))
}

type matrixCell uint32

// Empty reports whether this cell is empty.
func (c matrixCell) Empty() bool {
	return c == 0
}

// Source returns the register number containing this cell's input.
func (c matrixCell) Source() int {
	return int(c >> 24)
}

// Scale returns the factor to scale the contents of the source register by.
func (c matrixCell) Scale() Signal {
	return Signal(c << 8)
}

type Matrix struct {
	m *Memory
}

func (m *Memory) Matrix() Matrix {
	return Matrix{m}
}

const (
	NumMatrixRows     = 64
	MatrixCellsPerRow = 4 // it's a sparse matrix...

	MatrixBiasBase   = 0x40
	MatrixOutputBase = 0xc0
	MatrixCellsBase  = 0x100
)

// Reset restores the matrix to the default set of connections.
func (mm Matrix) Reset() {
	mm.Clear()

	// Set each output's first input to 100% of its bias.
	for i := range NumMatrixRows {
		mm.Connect(MatrixOutputBase+i, MatrixBiasBase+i, MaxSignal)
	}
}

// Clear restores the matrix to a state of no connectivity.
func (mm Matrix) Clear() {
	for i := range NumMatrixRows * MatrixCellsPerRow {
		r := mm.m.Register(MatrixCellsBase + i)
		r.Store(0)
	}
}

func (mm Matrix) Connect(dst, src int, gain Signal) {
	row := dst - MatrixOutputBase
	cellsBase := MatrixCellsBase + row*MatrixCellsPerRow

	for i := range MatrixCellsPerRow {
		r := mm.m.Register(cellsBase + i)

		c := matrixCell(r.load())
		if !c.Empty() && c.Source() != src {
			continue
		}

		r.Store((uint32(src) << 24) | (uint32(gain) >> 8))
		return
	}
	println("warning: Matrix.Connect: row", dst, "full")
}
