//go:build !tinygo

package keyboard

import "time"

type keyboard struct {
	rows [5]bool
	cols [4]bool

	rowPins []Row
	colPins []Column

	arp arpeggiator
}

func newKeyboard() Keyboard {
	kb := &keyboard{
		arp: arpeggiator{
			Notes: []uint8{48, 52, 55, 59, 60, 59, 55, 52},
			Tempo: 180 * time.Millisecond,
		},
	}

	kb.rowPins = make([]Row, len(kb.rows))
	kb.colPins = make([]Column, len(kb.cols))

	for i := range kb.rows {
		kb.rowPins[i] = &rowPin{kb, i}
	}
	for j := range kb.cols {
		kb.colPins[j] = &colPin{kb, j}
	}
	return kb
}

func (kb *keyboard) Rows() []Row {
	return kb.rowPins
}

func (kb *keyboard) Columns() []Column {
	return kb.colPins
}

type pin struct {
	kb  *keyboard
	num int
}

type rowPin pin
type colPin pin

func (p *rowPin) Set(level bool) {
	p.kb.rows[p.num] = level
	p.kb.update()
}

func (p *colPin) Get() bool {
	return p.kb.cols[p.num]
}

func (kb *keyboard) update() {
	kb.arp.Step()
	midinote := kb.arp.Note()      // 48..60
	scancode := int(midinote) - 41 //  7..19

	// clear columns
	for j := range kb.cols {
		kb.cols[j] = false
	}

	row := scancode / len(kb.cols)
	if !kb.rows[row] {
		return // no columns high
	}

	kb.cols[scancode%len(kb.cols)] = true
}

type arpeggiator struct {
	Notes       []uint8
	Tempo       time.Duration
	note        int
	lastStepped time.Time
}

func (arp *arpeggiator) Step() {
	now := time.Now()
	if now.Sub(arp.lastStepped) < arp.Tempo {
		return
	} else if !arp.lastStepped.IsZero() {
		arp.note = (arp.note + 1) % len(arp.Notes)
	}
	arp.lastStepped = now
}

func (arp *arpeggiator) Note() uint8 {
	return arp.Notes[arp.note]
}
