package microfont

import "gbenson.net/go/microfont"

func init() {
	f := microfont.Face04B08

	// Make exclamation mark narrower.
	if i, ok := microfont.GlyphIndex(f.Ranges, '!'); ok {
		f.Glyphs[i] >>= 1
	}

	// Shrink whitespace after ampersand.
	f.Kernings["& "] = -1
}

// Render renders s into buf, returning the number of bytes written.
func Render(buf []byte, s string) (n int) {
	f := microfont.Face04B08
	var lastr rune
	for _, r := range s {
		// Space the glyphs as necessary.
		if lastr != rune(0) {
			for _ = range kern(lastr, r) {
				buf[n] = 0
				n++
			}
		}
		lastr = r

		// Unpack the glyph into the buffer.
		for g, _ := f.GlyphFor(r); g > 0; g >>= 5 {
			buf[n] = byte(g & 0x1f)
			n++
		}
	}
	return n
}

// kern returns the inter-glyph spacing between two runes.
func kern(r0, r1 rune) int {
	f := microfont.Face04B08
	if r0 == ' ' {
		return 2
	}
	return f.Kern(r0, r1).Floor() + 1
}
