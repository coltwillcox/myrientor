package main

import (
	"os"
	"syscall"
	"time"
	"unsafe"
)

// listenForDrain starts listening for the drain hotkey ('q' / 'Q').
// When pressed, the returned channel receives once, signalling that no new
// files should be queued; active downloads are allowed to finish.
//
// The returned wait function blocks until the goroutine has exited and the
// terminal has been restored. Call it before printing final output to avoid
// garbled display.
func listenForDrain(done <-chan struct{}) (<-chan struct{}, func()) {
	drain := make(chan struct{}, 1)
	finished := make(chan struct{})

	go func() {
		defer close(finished)

		// Save original terminal state. If stdin is not a terminal, bail out.
		var orig syscall.Termios
		if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
			uintptr(syscall.Stdin),
			syscall.TIOCGETA,
			uintptr(unsafe.Pointer(&orig))); errno != 0 {
			return
		}

		// Enable raw mode: no echo, no line buffering.
		// ISIG is intentionally kept so Ctrl-C still sends SIGINT.
		raw := orig
		raw.Lflag &^= syscall.ICANON | syscall.ECHO
		syscall.Syscall(syscall.SYS_IOCTL,
			uintptr(syscall.Stdin),
			syscall.TIOCSETA,
			uintptr(unsafe.Pointer(&raw)))
		defer syscall.Syscall(syscall.SYS_IOCTL,
			uintptr(syscall.Stdin),
			syscall.TIOCSETA,
			uintptr(unsafe.Pointer(&orig)))

		// Make stdin non-blocking so the reader goroutine can exit cleanly
		// when this goroutine is done (important for multi-device syncs).
		syscall.SetNonblock(syscall.Stdin, true)
		defer syscall.SetNonblock(syscall.Stdin, false)

		innerDone := make(chan struct{})
		keyCh := make(chan byte, 4)
		go func() {
			buf := make([]byte, 1)
			for {
				n, err := os.Stdin.Read(buf)
				if n > 0 {
					select {
					case keyCh <- buf[0]:
					default:
					}
				}
				if err != nil {
					// EAGAIN (non-blocking, no data yet) or real error.
					// Poll until data arrives or we are asked to stop.
					select {
					case <-innerDone:
						return
					case <-time.After(50 * time.Millisecond):
					}
				}
			}
		}()

		for {
			select {
			case <-done:
				close(innerDone)
				return
			case key := <-keyCh:
				if key == 'q' || key == 'Q' {
					select {
					case drain <- struct{}{}:
					default:
					}
				}
			}
		}
	}()

	return drain, func() { <-finished }
}
