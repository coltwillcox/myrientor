package main

import (
	"fmt"
	"sync"
	"time"
)

type SyncStats struct {
	mu               sync.Mutex
	filesChecked     int
	filesDownloaded  int
	filesDeleted     int
	filesSkipped     int
	bytesDownloaded  int64                // Completed downloads total
	bytesInProgress  [maxConcurrent]int64 // Current progress per slot
	totalBytes       int64
	startTime        time.Time
	activities       [maxConcurrent]string // Track current activity in each slot
	activeSlots      int                   // Number of activity slots to display (min of maxConcurrent and file count)
	lastPrintedLines int                   // Number of lines printed in last Print() call (for cursor positioning)
}

func (s *SyncStats) IncrementChecked() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.filesChecked++
}

func (s *SyncStats) IncrementDownloaded(bytes int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.filesDownloaded++
	s.bytesDownloaded += bytes
}

func (s *SyncStats) IncrementDeleted() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.filesDeleted++
}

func (s *SyncStats) IncrementSkipped() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.filesSkipped++
}

func (s *SyncStats) SetTotalBytes(bytes int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.totalBytes = bytes
}

func (s *SyncStats) AddTotalBytes(bytes int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.totalBytes += bytes
}

func (s *SyncStats) SetSlotProgress(slot int, bytes int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if slot >= 0 && slot < maxConcurrent {
		s.bytesInProgress[slot] = bytes
	}
}

func (s *SyncStats) ClearSlotProgress(slot int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if slot >= 0 && slot < maxConcurrent {
		s.bytesInProgress[slot] = 0
	}
}

func (s *SyncStats) getTotalBytesTransferred() int64 {
	// Must be called with lock held
	total := s.bytesDownloaded
	for i := range maxConcurrent {
		total += s.bytesInProgress[i]
	}
	return total
}

func (s *SyncStats) SetActivity(slot int, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if slot >= 0 && slot < maxConcurrent {
		s.activities[slot] = message
	}
}

func (s *SyncStats) ClearActivity(slot int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if slot >= 0 && slot < maxConcurrent {
		s.activities[slot] = ""
	}
}

func (s *SyncStats) Print() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Count how many active activity lines we have
	activeCount := 0
	for i := range s.activeSlots {
		if s.activities[i] != "" {
			activeCount++
		}
	}

	// Calculate lines to print this time (activity lines + 3 stats lines)
	linesToPrint := activeCount + 3

	// Move cursor up to overwrite previous lines
	for i := range s.lastPrintedLines {
		if i > 0 {
			fmt.Print("\033[A") // Move cursor up one line
		}
		fmt.Print("\r\033[K") // Clear entire line
	}

	// Print only non-empty activity lines
	for i := range s.activeSlots {
		if s.activities[i] != "" {
			fmt.Printf("%s\n", s.activities[i])
		}
	}

	// Remember how many lines we printed for next time
	s.lastPrintedLines = linesToPrint

	// Calculate stats using real-time bytes (completed + in-progress)
	totalTransferred := s.getTotalBytesTransferred()
	elapsed := time.Since(s.startTime)
	speed := int64(0)
	if elapsed.Seconds() > 0 {
		speed = int64(float64(totalTransferred) / elapsed.Seconds())
	}

	// Calculate percentage if total is known
	progressStr := ""
	if s.totalBytes > 0 {
		percentage := float64(totalTransferred) / float64(s.totalBytes) * 100
		progressStr = fmt.Sprintf(" (%.1f%%)", percentage)
	}

	// Print stats on separate lines
	fmt.Printf("%sFiles:%s %d checked, %s%d downloaded%s, %d skipped, %s%d deleted%s\n",
		colorBold,
		colorReset,
		s.filesChecked,
		colorGreen,
		s.filesDownloaded,
		colorReset,
		s.filesSkipped,
		colorYellow,
		s.filesDeleted,
		colorReset)

	fmt.Printf("%sTransfer:%s %s%s%s / %s%s @ %s%s/s%s\n",
		colorBold,
		colorReset,
		colorCyan,
		formatBytes(totalTransferred),
		colorReset,
		formatBytesIfKnown(s.totalBytes),
		progressStr,
		colorCyan,
		formatBytes(speed),
		colorReset)

	fmt.Printf("%sTime:%s %s%s%s",
		colorBold,
		colorReset,
		colorBlue,
		formatDuration(elapsed),
		colorReset)
}
