package fmt

// As of 20260602, using real fmt.Sprint more than quadruples the
// UF2 payload size, from 51 to 222 blocks (13056 to 56832 bytes).
// That extra 43 KiB is 2% of a standard Pico's flash, which is a
// big chunk for something thats only current user is a panic handler
// that ideally shouldn't ever run.

func Sprint(args ...any) string {
	var s string
	for _, a := range args {
		switch arg := a.(type) {
		case string:
			s += arg
		case error:
			s += arg.Error()
		default:
			println("panic: ", arg, "not formattable")
			s += "<not-formattable>"
		}
	}
	return s
}
