//go:build !tinygo

package keyboard

type pin uint8

func (p pin) Get() bool  { return false }
func (p pin) Set(v bool) {}

var rowPins [7]pin
var colPins [8]pin

type keyboard struct {
	rows []Row
	cols []Column
}

func newKeyboard() Keyboard {
	kb := &keyboard{
		make([]Row, len(rowPins)),
		make([]Column, len(colPins)),
	}
	for i, row := range rowPins {
		kb.rows[i] = row
	}
	for j, col := range colPins {
		kb.cols[j] = col
	}
	return kb
}

func (kb *keyboard) Rows() []Row {
	return kb.rows
}

func (kb *keyboard) Columns() []Column {
	return kb.cols
}
