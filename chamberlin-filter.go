package picosynth

// ChamberlinFilter is a state variable filter.
//
// Sources:
//   - Hal Chamberlin, “Musical Applications of Microprocessors,”
//     (2nd ed.), Hayden Book Company 1985. pp 490-492.
//     (via https://www.musicdsp.org/en/latest/Filters/142-state-variable-filter-chamberlin-version.html)
//   - https://dsp.stackexchange.com/questions/70939/frequency-response-of-a-digital-state-variable-chamberlin-filter (has diagram)
//   - https://en.wikipedia.org/wiki/State_variable_filter (background info)
type ChamberlinFilter struct {
	Frequency Frequency // center frequency
	Resonance Signal    // resonance

	Input Signal // input

	D1 Int4_28 // delay associated with bandpass output
	D2 Int4_28 // delay associated with low-pass output

	Lout Signal // low-pass output
	Bout Signal // high-pass output
	Hout Signal // bandpass output
	Nout Signal // notch output
}

func (f *ChamberlinFilter) Step() {
	// Parameters -------------------------------------------------

	// Chamberlin gives two options for calculating the frequency
	// control parameter F1:
	//
	// | simple frequency tuning with error towards nyquist:
	// | F1 = 2*pi*F/Fs
	// |
	// | ideal tuning:
	// | F1 = 2*sin(pi*F/Fs)
	//
	// Where F is the desired center frequency, and Fs the sample
	// rate.  We can do sin easily enough, lets use the ideal one.
	//
	// What's F/Fs?
	//  | F     | F/Fs | pi*F/Fs
	//  | ----- | ---- | -------
	//  | 0     | 0    | 0
	//  | Fs/2  | 0.5  | π/2
	//
	// Scaling:
	//  - `Frequency(MaxSignal+1)` ≡ Fs/2 (nyquist),
	//  - `Signal.Sin(MaxSignal)` is interpreted as sin(π),
	//  - so sin(pi*F/Fs) is `Signal(Frequency(f)/2).Sin()`.
	//
	// But then, Signal.Sin(π/2 equivalent) returns MaxInt32;
	// there's no scope to double it so we have 2*sin(pi*F/Fs).
	// I ran a hillclimber on a float64 version of this filter
	// to find maximum absolute value of any output when fed a
	// MaxSignal step input and it found 4.88903 (F=17027.8Hz,
	// Q1=0.214497), so we need to use signed Q4.28 fixed point
	// math to represent that.
	q230_F1 := Int2_30(Signal(f.Frequency / 2).Sin())
	// Q2.30 because Q2.30(x) has the same bits as Q1.31(x/2),
	// so casting Q1.31 to Q2.30 is essentially a doubling of
	// the represented value so what the Q2.30 valuerepresents
	// is _2_*sin(pi*F/Fs), the value we're after :)

	// Chamberlin's Q control parameter is the inverse of Q, so
	// Q1 goes from 2 to 0 as Q goes from 0.5 to infinity, but
	// I don't want to divide, so I'm going to vary it linearly
	// such that:
	//   - f.Resonance=Signal(0) [  0%] -> Q1=2
	//   - f.Resonance=MaxSignal [100%] -> Q1=0
	//
	// (Combining Q1=1/Q [Chamberlin] and Q1=2*(1-f.Resonance) [us]
	// means f.Resonance = 1-1/(2Q), and Q=1/(2*(1-f.Resonance)),
	// but that's not important.)
	q230_Q1 := Int2_30(MaxSignal - f.Resonance)
	// [Again casting Q1.31 to Q2.30 for that sweet implicit doubling]

	F1 := q230_F1.Int4_28() // Frequency control parameter
	Q1 := q230_Q1.Int4_28() // Q control parameter

	// Inputs -----------------------------------------------------
	I := f.Input.Int4_28() // input sample

	// State --------------------------------------------
	D1 := f.D1 // delay associated with bandpass output
	D2 := f.D2 // delay associated with low-pass output

	// Algorithm --------------------------------------------------
	L := D2 + F1.ClampedMul(D1)    // low-pass output sample
	H := I - L - Q1.ClampedMul(D1) // high-pass output sample
	B := F1.ClampedMul(H) + D1     // bandpass output sample
	N := H + L                     // notch output sample

	// Store state ------------------------------------------------
	f.D1 = B
	f.D2 = L

	// Store outputs ----------------------------------------------
	// XXX we cast, this is like dividing by 8, what to do... compress?
	f.Lout = Signal(L)
	f.Bout = Signal(B)
	f.Hout = Signal(H)
	f.Nout = Signal(N)
}
