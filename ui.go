package picosynth

type UI interface {
	// PageHasFocus reports whether p has focus.
	PageHasFocus(p Page) bool

	// YieldFocus yields focus to the default page.
	YieldFocus()

	// InvalidateDisplay requests a redraw at the next refresh.
	InvalidateDisplay()

	// ScreenBlanked reports whether the screen is blanked.
	ScreenBlanked() bool
}
