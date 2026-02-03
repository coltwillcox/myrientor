package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// ErrorLogger handles writing errors to a log file
type ErrorLogger struct {
	mu       sync.Mutex
	file     *os.File
	filename string
	count    int
}

// NewErrorLogger creates a new error logger with timestamped filename
// File is opened lazily on first Log call
func NewErrorLogger() *ErrorLogger {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("myrientor-errors_%s.log", timestamp)

	return &ErrorLogger{
		filename: filename,
	}
}

// Log writes an error to the log file
// Opens the file lazily on first call
func (l *ErrorLogger) Log(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.count++

	// Open file lazily on first log
	if l.file == nil {
		file, err := os.OpenFile(l.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return // Silently discard if file can't be opened
		}
		l.file = file
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(l.file, "[%s] %s\n", timestamp, message)
}

// Close closes the log file if it was opened
func (l *ErrorLogger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		l.file.Close()
	}
}

// Count returns the number of errors logged
func (l *ErrorLogger) Count() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.count
}

// Filename returns the log filename
func (l *ErrorLogger) Filename() string {
	return l.filename
}
