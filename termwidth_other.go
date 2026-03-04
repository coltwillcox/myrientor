//go:build !linux && !darwin && !windows

package main

// terminalWidth returns a conservative default on platforms where we don't
// query the terminal size.
func terminalWidth() int {
	return 80
}
