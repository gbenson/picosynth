package picosynth

import "gbenson.net/go/picosynth/internal/ui"

//go:generate python3 make-register-table.py

type Register struct {
	m *Memory
	n int
}

type RegisterInfo interface {
	// Name returns the name of the register, or the empty string if
	// the register is unassigned.
	Name() string

	// Parameter returns a [ui.Parameter] representing this register.
	Parameter(n int) ui.Parameter
}

// Info returns information about the register, or nil if the register
// is unassigned.
func (r *Register) Info() RegisterInfo {
	return registerTable[r.n]
}

// Assigned reports whether the register has been assigned a role.
// Unassigned registers are not currently used by Picosynth.
func (r *Register) Assigned() bool {
	return r.Info() != nil
}

// Name returns the name of the register, or the empty string if the
// register is unassigned.
func (r *Register) Name() string {
	if ri := r.Info(); ri != nil {
		return ri.Name()
	}
	return ""
}

// Editable reports whether the contents of r should be editable in
// the memory editor.  All registers are editable in theory, but in
// practice it makes no sense to edit feedback registers or matrix
// outputs since their contents will be replaced for every sample.
func (r *Register) Editable() bool {
	n := r.n
	switch {
	case n&0x180 == 0x80:
		return false
	case n == Filt1Cutoff:
		return false
	case n == Filt1Resonance:
		return false
	default:
		return true
	}
}

func (r *Register) Load() uint32 {
	return r.m.Load(r.n)
}

func (r *Register) load() uint32 {
	return r.m.load(r.n)
}

func (r *Register) Store(v uint32) {
	r.m.Store(r.n, v)
}

// A RegisterParameter wraps a register.
type RegisterParameter int

// Parameter implements [ParameterSpec].
func (p RegisterParameter) Parameter() ui.Parameter {
	r := Register{nil, int(p)}
	return r.parameter()
}

func (r *Register) parameter() ui.Parameter {
	return r.Info().Parameter(r.n)
}

// Generic register types.
type (
	NonEditableRegister string // [not editable]
	SignalRegister      string // Signal, full-range [MinSignal,MaxSignal]
	MatrixCellRegister  string // split (src/amount; see XXX note below)
)

// Specific register types.
type (
	FilterModeRegister string // cycle through values

	DetuneRegister   = SignalRegister // ±Pitch, prob range-limited?
	DurationRegister = SignalRegister // positive time
	GainRegister     = SignalRegister // Signal, full-range
	PhaseRegister    = SignalRegister // Signal, full-range
	RateRegister     = SignalRegister // Frequency?
)

func (r NonEditableRegister) Name() string {
	return string(r)
}

func (r NonEditableRegister) Parameter(n int) ui.Parameter {
	panic("should not call")
}

func (r SignalRegister) Name() string {
	return string(r)
}

func (r SignalRegister) Parameter(n int) ui.Parameter {
	return &ui.NumericParameter[Signal]{
		Name:     r.Name(),
		Register: n,
		Min:      MinSignal,
		Max:      MaxSignal,
	}
}

func (r FilterModeRegister) Name() string {
	return string(r)
}

func (r FilterModeRegister) Parameter(n int) ui.Parameter {
	return &ui.EnumParameter{
		Name:     r.Name(),
		Register: n,
		// XXX should have something in filter-mode.go to generate these
		Names: []string{
			"bypass",
			"low pass",
			"high pass",
			"band pass",
			"notch",
		},
	}
}

func (r MatrixCellRegister) Name() string {
	return string(r)
}

func (r MatrixCellRegister) Parameter(n int) ui.Parameter {
	// XXX when adding the matrix parameter group,
	// (or just when adding parameters which are matrix elements)
	// add MatrixCell{Source,Amount}Parameter types to go with
	// the generic RegisterParameter type, and have those
	// generate an EnumParameter with all the sources and a
	// numericparameter with 24-bit min,max; and either extend
	// those classes to allow specifying a bitfield mask and
	// shift (probably), or (likely less good) make special
	// matrixcell{source,amount}parameter classes in intrnal/ui
	panic("should not call")
}
