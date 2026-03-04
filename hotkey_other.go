//go:build !linux

package main

// listenForDrain is a no-op stub on non-Linux platforms.
func listenForDrain(_ <-chan struct{}) (<-chan struct{}, func()) {
	return make(chan struct{}), func() {}
}
