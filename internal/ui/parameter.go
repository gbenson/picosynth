package ui

import "sync/atomic"

type Parameter interface {
	Focus(m Memory)
	Adjust(m Memory, amount int32)
	Render(d *Display)
}

type Value interface {
	~int32 | ~uint32
}

type NumericParameter[T Value] struct {
	Name     string
	Register int
	Min, Max T

	value atomic.Uint32
}

func (p *NumericParameter[T]) Focus(m Memory) {
	p.update(T(m.Load(p.Register)))
}

func (p *NumericParameter[T]) Adjust(m Memory, amount int32) {
	step := int32((int64(p.Max) - int64(p.Min)) >> 7) // XXX accelerate
	amount *= step

	r := p.Register
	v := clampedAdd(T(m.Load(r)), amount, p.Min, p.Max)
	m.Store(r, uint32(v))
	p.update(v)
}

func (p *NumericParameter[T]) update(v T) {
	p.value.Store(uint32(v))
}

func (p *NumericParameter[T]) Render(d *Display) {
	d.Clear()
	RenderRegisterName(d, p.Name)
	RenderHexValue(d, p.value.Load())
	d.Sync()
}

type EnumParameter struct {
	Name     string
	Register int
	Names    []string

	value atomic.Uint32
}

func (p *EnumParameter) Focus(m Memory) {
	p.update(m.Load(p.Register))
}

func (p *EnumParameter) Adjust(m Memory, amount int32) {
	r := p.Register
	v := wrappedAdd(m.Load(r), amount, 0, uint32(len(p.Names)-1))
	m.Store(r, v)
	p.update(v)
}

func (p *EnumParameter) update(v uint32) {
	p.value.Store(v)
}

func (p *EnumParameter) Render(d *Display) {
	d.Clear()
	RenderRegisterName(d, p.Name)
	s := "bork"
	i := p.value.Load()
	if i < uint32(len(p.Names)) {
		s = p.Names[i]
	}
	RenderTextValue(d, s)
	d.Sync()
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
