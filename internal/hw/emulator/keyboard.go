package emulator

import (
	"sync"

	"github.com/veandco/go-sdl2/sdl"
)

//go:generate python3 make-keyboard-layout.py

const (
	KeyStrideX = WindowWidth / 15
	KeyMarginX = KeyStrideX / 20
	KeyWidth   = KeyStrideX - KeyMarginX

	KeyHeight  = KeyWidth + 2*KeyMarginX // slightly taller than wider
	KeyMarginY = KeyHeight / 8
	KeyStrideY = KeyHeight + KeyMarginY

	KeyCornerRadius = KeyMarginX * 2

	whiteKeyRowWidth  = KeyStrideX*12 - KeyMarginX
	whiteKeyRowOffset = (WindowWidth-whiteKeyRowWidth)/2 - KeyWidth/4
)

var keyRowOffsets = []int32{
	whiteKeyRowOffset - KeyWidth*5/6,
	whiteKeyRowOffset + KeyWidth/2,
	whiteKeyRowOffset,
	whiteKeyRowOffset + KeyWidth/2,
}

type Keyboard struct {
	mu     sync.Mutex
	matrix [8][7]bool
	keymap Keymap
}

type Keymap map[sdl.Keycode]*SoftKey

var keyboard Keyboard

func (kb *Keyboard) matrixSize() (outs, ins int) {
	return len(kb.matrix[0]), len(kb.matrix)
}

func (kb *Keyboard) ensureKeymap() {
	if kb.keymap != nil {
		return // already built
	}

	kb.keymap = make(Keymap)
	for _, row := range keyboardLayout {
		for _, key := range row {
			kb.keymap[key.Code] = &key
		}
	}
}

func (kb *Keyboard) handleEvent(e *sdl.KeyboardEvent) {
	if e.Repeat != 0 {
		return
	}

	kb.mu.Lock()
	defer kb.mu.Unlock()
	kb.ensureKeymap()

	key := kb.keymap[e.Keysym.Sym]
	if key == nil {
		return
	}

	var held bool
	switch e.Type {
	case sdl.KEYDOWN:
		held = true
	case sdl.KEYUP:
	default:
		return
	}

	kb.matrix[key.KI][key.KO] = held
}

func (kb *Keyboard) Render(draw Draw) {
	kb.mu.Lock()
	defer kb.mu.Unlock()

	draw.Box(0, 0, WindowWidth, WindowHeight, Gray(47))

	const y0 = KeyHeight
	for row := len(keyboardLayout) - 1; row >= 0; row-- {
		y := int32(y0 + row*KeyStrideY)
		h := int32(KeyHeight)
		if row == 2 {
			// white keys
			y -= KeyStrideY
			h += KeyStrideY
		}

		x := keyRowOffsets[row]
		for _, key := range keyboardLayout[row] {
			x += KeyStrideX * int32(key.Offset)
			w := int32(KeyWidth)
			if key.Wide {
				w += KeyWidth / 2
				x -= KeyWidth / 3
				x += KeyMarginX // fudge
			}

			bgColor := key.Color
			fgColor := Black

			held := kb.matrix[key.KI][key.KO]
			if held {
				bgColor = Pink
				fgColor = White
			} else if row == 1 {
				// black keys
				fgColor = White
			}

			draw.RoundedBox(x, y, w, h, KeyCornerRadius, bgColor)

			draw.Text(
				x+KeyCornerRadius+1,
				y+h-KeyCornerRadius-16,
				key.Name,
				fgColor,
			)

			x += KeyStrideX
		}
	}
}
