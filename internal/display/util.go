package display

func i2s(i int32) string {
	return string(rune(i))
}

func r2i(r rune) int32 {
	return int32(r)
}
