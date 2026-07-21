//go:build !pico

package hw

type Pin interface {
	Get() bool
	Set(bool)
}
