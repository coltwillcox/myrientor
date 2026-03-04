//go:build linux || darwin

package main

import (
	"syscall"
	"unsafe"
)

// winsize mirrors the kernel's struct winsize (not available in stdlib syscall).
type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

// terminalWidth returns the current terminal width in columns, or 80 if unknown.
func terminalWidth() int {
	var ws winsize
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdout),
		syscall.TIOCGWINSZ,
		uintptr(unsafe.Pointer(&ws))); errno == 0 && ws.Col > 0 {
		return int(ws.Col)
	}
	return 80
}
