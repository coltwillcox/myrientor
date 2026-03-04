package main

import (
	"syscall"
	"unsafe"
)

const (
	stdInputHandle  = uintptr(0xFFFFFFF6) // (DWORD)(-10) = STD_INPUT_HANDLE
	enableLineInput = 0x0002
	enableEchoInput = 0x0004
	waitObject0     = 0x00000000
	keyEvent        = 0x0001
)

var (
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	procGetStdHandle        = kernel32.NewProc("GetStdHandle")
	procGetConsoleMode      = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode      = kernel32.NewProc("SetConsoleMode")
	procWaitForSingleObject = kernel32.NewProc("WaitForSingleObject")
	procReadConsoleInputW   = kernel32.NewProc("ReadConsoleInputW")
)

// inputRecord matches the Win32 INPUT_RECORD layout (20 bytes).
// The Event union is overlaid with KEY_EVENT_RECORD fields.
type inputRecord struct {
	EventType uint16
	_         [2]byte // padding to align union to 4 bytes
	KeyDown   int32   // BOOL bKeyDown
	Repeat    uint16  // WORD wRepeatCount
	VK        uint16  // WORD wVirtualKeyCode
	VSC       uint16  // WORD wVirtualScanCode
	Char      uint16  // WCHAR UnicodeChar
	Control   uint32  // DWORD dwControlKeyState
}

// listenForDrain starts listening for the drain hotkey ('q' / 'Q').
// When pressed, the returned channel receives once, signalling that no new
// files should be queued; active downloads are allowed to finish.
//
// The returned wait function blocks until the goroutine has exited and the
// console mode has been restored. Call it before printing final output to
// avoid garbled display.
func listenForDrain(done <-chan struct{}) (<-chan struct{}, func()) {
	drain := make(chan struct{}, 1)
	finished := make(chan struct{})

	go func() {
		defer close(finished)

		// Get the stdin console handle.
		hStdin, _, _ := procGetStdHandle.Call(stdInputHandle)
		if hStdin == 0 || hStdin == ^uintptr(0) {
			return
		}

		// Save original console mode. If stdin is not a console (e.g.
		// redirected from a file), GetConsoleMode returns 0 and we bail out.
		var origMode uint32
		ret, _, _ := procGetConsoleMode.Call(hStdin, uintptr(unsafe.Pointer(&origMode)))
		if ret == 0 {
			return
		}

		// Disable line input and echo; keep ENABLE_PROCESSED_INPUT so
		// Ctrl-C still generates a CTRL_C_EVENT / SIGINT.
		rawMode := origMode &^ uint32(enableLineInput|enableEchoInput)
		procSetConsoleMode.Call(hStdin, uintptr(rawMode))
		defer procSetConsoleMode.Call(hStdin, uintptr(origMode))

		for {
			// Check for shutdown before blocking.
			select {
			case <-done:
				return
			default:
			}

			// Wait up to 50 ms for input to appear in the console buffer.
			ret, _, _ := procWaitForSingleObject.Call(hStdin, 50)
			if ret != waitObject0 {
				// Timeout or error — loop back to check done.
				continue
			}

			// Read one input record from the console buffer.
			var rec inputRecord
			var numRead uint32
			procReadConsoleInputW.Call(
				hStdin,
				uintptr(unsafe.Pointer(&rec)),
				1,
				uintptr(unsafe.Pointer(&numRead)),
			)
			if numRead == 0 || rec.EventType != keyEvent || rec.KeyDown == 0 {
				continue
			}
			if rec.Char == 'q' || rec.Char == 'Q' {
				select {
				case drain <- struct{}{}:
				default:
				}
			}
		}
	}()

	return drain, func() { <-finished }
}
