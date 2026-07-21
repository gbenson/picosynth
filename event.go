package picosynth

import "gbenson.net/go/picosynth/internal/fmt"

type Event uint

const (
	eventDataMask = Event((1 << 24) - 1)
	eventTypeMask = ^eventDataMask

	EventTypeKeyDown Event = (eventDataMask + 1) << iota
	EventTypeKeyUp
	EventTypeADC
	EventTypeEncoder
)

func NewKeyDownEvent(c Scancode) Event {
	return EventTypeKeyDown | Event(c)
}

func NewKeyUpEvent(c Scancode) Event {
	return EventTypeKeyUp | Event(c)
}

func NewADCEvent(n uint8, v uint16) Event {
	return EventTypeADC | (Event(n) << 16) | Event(v)
}

func NewEncoderEvent(n uint8, d int16) Event {
	return EventTypeEncoder | (Event(n) << 16) | (Event(d) & 0xffff)
}

func (e Event) Type() Event {
	return e & eventTypeMask
}

func (e Event) Number() uint8 {
	return uint8(e >> 16)
}

func (e Event) Value() uint16 {
	return uint16(e)
}

func (e Event) Delta() int16 {
	return int16(e.Value())
}

func (e Event) Scancode() Scancode {
	return Scancode(e & eventDataMask)
}

// String implements [fmt.Stringer].
func (e Event) String() string {
	return fmt.Hex32Stringer("Event", e)
}
