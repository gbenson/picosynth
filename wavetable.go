package picosynth

//go:generate go run make-sine-table.go

const (
	WavetableIndexBits = 6
	WavetableSize      = 1 << WavetableIndexBits

	// 6 bits => 64 slots => 256 bytes/wavetable

	wtIndexShift    = SignalBits - WavetableIndexBits
	wtIndexMask     = WavetableSize - 1
	wtFractionMask  = (1 << wtIndexShift) - 1
	wtFractionShift = WavetableIndexBits - 1
)

// A Wavetable samples one cycle of a periodic waveform.
type Wavetable [WavetableSize]Signal

// Get returns the amplitude of the waveform at phase x.
func (wt *Wavetable) Get(x Signal) Signal {
	i := int(uint32(x)) >> wtIndexShift
	j := (i + 1) & wtIndexMask

	yi := wt[i]
	yj := wt[j]

	// y = mx + c
	c := yi
	m := yj - yi

	return m.Mul((x&wtFractionMask)<<wtFractionShift) + c
}
