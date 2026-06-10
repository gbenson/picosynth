package keyboard

type Keyboard interface {
	Rows() []Row
	Columns() []Column
}

type Row interface {
	Set(bool)
}

type Column interface {
	Get() bool
}

func New() Keyboard {
	return newKeyboard()
}
