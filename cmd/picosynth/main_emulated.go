//go:build !baremetal

package main

import (
	"fmt"
	"os"

	"gbenson.net/go/picosynth"
)

func main() {
	if err := picosynth.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
