package picosynth

//go:generate python3 make-register-table.py

type Register struct {
	m *Memory
	n int
}

type RegisterInfo interface {
	// Name returns the name of the register, or the empty string if
	// the register is unassigned.
	Name() string
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

// Generic register types.
type (
	NonEditableRegister string // [not editable]
	SignalRegister      string // Signal, full-range [MinSignal,MaxSignal]
)

// Specific register types.
type (
	DetuneRegister     = SignalRegister // ±Pitch, prob range-limited?
	DurationRegister   = SignalRegister // positive time
	FilterModeRegister = SignalRegister // cycle through values
	GainRegister       = SignalRegister // Signal, full-range
	MatrixCellRegister = SignalRegister // split (src/amount)
	PhaseRegister      = SignalRegister // Signal, full-range
	RateRegister       = SignalRegister // Frequency?
)

func (ri NonEditableRegister) Name() string {
	return string(ri)
}

func (ri SignalRegister) Name() string {
	return string(ri)
}
