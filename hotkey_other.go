//go:build !linux && !darwin && !windows

package main

// listenForDrain is a no-op stub on unsupported platforms.
func listenForDrain(_ <-chan struct{}) (<-chan struct{}, func()) {
	return make(chan struct{}), func() {}
}
