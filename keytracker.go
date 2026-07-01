package picosynth

type KeyTracker struct {
	Transpose int // Transposition to apply, in semitones.

	notes [128]Note // Which playing note did each held-down key trigger?

	Note Note // The last played note.
	Gate bool // true if a note is playing, false otherwise.
}

func (kt *KeyTracker) init() {
	for key := range kt.notes {
		kt.notes[key] = NoNote
	}
}

func (kt *KeyTracker) Receive(key Note, down bool) {
	note := NoNote
	if down {
		note = key.Transpose(kt.Transpose)
	}
	kt.notes[key] = note
}

func (kt *KeyTracker) Step() {
	winner := NoNote

	for _, note := range kt.notes {
		if !note.IsValid() {
			continue
		} else if note > winner || winner == NoNote {
			// highest wins
			winner = note
		}
	}

	if winner.IsValid() {
		kt.Note = winner
		kt.Gate = true
	} else {
		kt.Gate = false
		// leave kt.Note floating
	}
}
