package ui

type Parameter interface {
	Adjust(system System, amount int32)
	Render(system System)
}

type Value interface {
	~int32 | ~uint32
}

type NumericParameter[T Value] struct {
	Name     string
	Register int
	Min, Max T
}

func (p *NumericParameter[T]) Adjust(system System, amount int32) {
	step := int32((int64(p.Max) - int64(p.Min)) >> 7) // XXX accelerate
	ClampedAdjust(system, p.Register, p.Min, p.Max, amount*step)
}

func (p *NumericParameter[T]) Render(system System) {
	d := system.Display()
	d.Clear()
	RenderRegisterName(d, p.Name)
	RenderHexValue(d, system.Load(p.Register))
	d.Sync()
}

type EnumParameter struct {
	Name     string
	Register int
	Names    []string
}

func (p *EnumParameter) Adjust(system System, amount int32) {
	WrappedAdjust(system, p.Register, 0, uint32(len(p.Names)-1), amount)
}

func (p *EnumParameter) Render(system System) {
	d := system.Display()
	d.Clear()
	RenderRegisterName(d, p.Name)
	s := "bork"
	i := system.Load(p.Register)
	if i < uint32(len(p.Names)) {
		s = p.Names[i]
	}
	RenderTextValue(d, s)
	d.Sync()
}

// ClampedAdjust adds amount to register r clamping the result to [minv,maxv].
func ClampedAdjust[T Value](system System, r int, minv, maxv T, amount int32) {
	system.Store(r, uint32(clampedAdd(T(system.Load(r)), amount, minv, maxv)))
}

// WrappedAdjust adds amount to register, wrapping the result at [minv,maxv].
func WrappedAdjust[T Value](system System, r int, minv, maxv T, amount int32) {
	system.Store(r, uint32(wrappedAdd(T(system.Load(r)), amount, minv, maxv)))
}

// clampedAdd returns x+y clamped to [minv,maxv].
func clampedAdd[T Value](x T, y int32, minv, maxv T) T {
	return T(max(int64(minv), min(int64(x)+int64(y), int64(maxv))))
}

// wrappedAdd returns x+y wrapped at [minv,maxv].
func wrappedAdd[T Value](x T, y int32, minv, maxv T) T {
	offset := int64(minv)

	sum := int64(x) + int64(y) - offset
	lim := int64(maxv) + 1 - offset
	if sum < 0 {
		for sum < 0 {
			sum += lim
		}
	} else {
		sum %= lim
	}

	return T(sum + offset)
}
