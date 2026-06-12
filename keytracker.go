package picosynth

type KeyTracker struct {
	notes [128]bool

	Note Note // The last played note.
	Gate bool // true if a note is playing, false otherwise.
}

func (kt *KeyTracker) Receive(key Note, down bool) {
	kt.notes[key] = down
}

func (kt *KeyTracker) Step() {
	for key, down := range kt.notes {
		if !down {
			continue
		}
		kt.Note = Note(key)
		kt.Gate = true
	}
}
