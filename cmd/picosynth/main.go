package main

import "gbenson.net/go/picosynth"

func main() {
	defer func() { fatal(recover()) }()

	var ps picosynth.Engine
	if err := ps.Run(); err != nil {
		fatal(err)
	}

	fatal("unexpected nil error")
}
