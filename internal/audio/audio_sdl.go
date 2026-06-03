//go:build !tinygo

package audio

import "errors"

func open(sampleRate int) (Device, error) {
	return nil, errors.New("not implemented")
}
