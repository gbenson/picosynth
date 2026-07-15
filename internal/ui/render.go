package ui

func RenderRegisterName(d *Display, name string) {
	d.TextAt(0, 0, 16, name)
}

func RenderHexValue(d *Display, v uint32) {
	for i := range 4 {
		RenderHexAt(d, int32(35*i+1), 16, 16, uint8((v>>((3-i)*8))&255))
	}
}

func RenderTextValue(d *Display, v string) {
	d.TextAt(0, 16, 16, v)
}

func RenderHexAt(d *Display, x, y, h int32, v uint8) {
	const digits = "0123456789abcdef"

	d.TextAt(x, y, h, string(digits[v>>4]))
	d.TextAt(x+(h/8)*6, y, h, string(digits[v&15]))
}
