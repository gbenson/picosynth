package picosynth

//go:generate python3 make-register-table.py

type Register struct {
	m *Memory
	n int
}

// Name returns the name of the register, or the empty string if
// the register is unassigned.
func (r *Register) Name() string {
	return RegisterNames[r.n]
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
