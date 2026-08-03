package main

import "gbenson.net/go/picosynth"

func main() {
	defer func() { fatal(recover()) }()

	if err := picosynth.Run(); err != nil {
		fatal(err)
	}

	fatal("unexpected nil error")
}
