package main

import (
	"unsafe"
)

const stdOutputHandle = uintptr(0xFFFFFFF5) // (DWORD)(-11) = STD_OUTPUT_HANDLE

var procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")

// The following types mirror the Win32 COORD, SMALL_RECT, and
// CONSOLE_SCREEN_BUFFER_INFO structures (22 bytes total).
type coord struct{ X, Y int16 }
type smallRect struct{ Left, Top, Right, Bottom int16 }
type consoleScreenBufferInfo struct {
	Size              coord    // offset 0  (4 bytes)
	CursorPosition    coord    // offset 4  (4 bytes)
	Attributes        uint16   // offset 8  (2 bytes)
	Window            smallRect // offset 10 (8 bytes)
	MaximumWindowSize coord    // offset 18 (4 bytes)
}

// terminalWidth returns the current console window width in columns, or 80
// if the query fails (e.g. stdout is redirected).
func terminalWidth() int {
	hStdout, _, _ := procGetStdHandle.Call(stdOutputHandle)
	if hStdout == 0 || hStdout == ^uintptr(0) {
		return 80
	}
	var info consoleScreenBufferInfo
	ret, _, _ := procGetConsoleScreenBufferInfo.Call(hStdout, uintptr(unsafe.Pointer(&info)))
	if ret == 0 {
		return 80
	}
	w := int(info.Window.Right-info.Window.Left) + 1
	if w <= 0 {
		return 80
	}
	return w
}
