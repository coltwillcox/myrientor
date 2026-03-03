package main

import (
	"fmt"
	"sync"
	"time"
)

type speedSample struct {
	t     time.Time
	bytes int64
}

type SyncStats struct {
	mu                     sync.Mutex
	filesTotal             int
	filesChecked           int
	filesDownloaded        int
	filesDeleted           int
	filesSkipped           int
	filesErrors            int
	bytesDownloaded        int64   // Completed bytes including skipped (for transfer progress display)
	bytesActuallyDownloaded int64  // Only actual downloads, for speed calculation
	bytesInProgress        []int64 // Current progress per slot
	slotBytesBase          []int64 // Cumulative completed bytes per slot (for monotonic speed samples)
	totalBytes             int64
	startTime              time.Time
	activities             []string       // Track current activity in each slot
	activeSlots            int            // Number of activity slots to display (min of maxConcurrent and file count)
	lastPrintedLines       int            // Number of lines printed in last Print() call (for cursor positioning)
	maxConcurrent          int            // Maximum concurrent downloads
	globalSpeedSamples     []speedSample  // Sliding window for global download speed
	slotSpeedSamples       [][]speedSample // Sliding window per slot for per-file speed
}

func NewSyncStats(maxConcurrent int) *SyncStats {
	return &SyncStats{
		startTime:        time.Now(),
		bytesInProgress:  make([]int64, maxConcurrent),
		slotBytesBase:    make([]int64, maxConcurrent),
		activities:       make([]string, maxConcurrent),
		slotSpeedSamples: make([][]speedSample, maxConcurrent),
		maxConcurrent:    maxConcurrent,
	}
}

func (s *SyncStats) SetFilesTotal(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.filesTotal = n
}

func (s *SyncStats) IncrementChecked() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.filesChecked++
}

func (s *SyncStats) IncrementDownloaded(slot int, bytes int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.filesDownloaded++
	s.bytesDownloaded += bytes
	s.bytesActuallyDownloaded += bytes
	if slot >= 0 && slot < s.maxConcurrent {
		s.slotBytesBase[slot] += bytes
	}
}

func (s *SyncStats) IncrementDeleted() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.filesDeleted++
}

func (s *SyncStats) IncrementSkipped(bytes int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.filesSkipped++
	s.bytesDownloaded += bytes
}

func (s *SyncStats) IncrementErrors() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.filesErrors++
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
	if slot >= 0 && slot < s.maxConcurrent {
		s.bytesInProgress[slot] = bytes
		// Record slot speed sample
		now := time.Now()
		cumulative := s.slotBytesBase[slot] + bytes
		s.slotSpeedSamples[slot] = append(s.slotSpeedSamples[slot], speedSample{t: now, bytes: cumulative})
		// Prune samples older than 12 seconds
		cutoff := now.Add(-12 * time.Second)
		for len(s.slotSpeedSamples[slot]) > 1 && s.slotSpeedSamples[slot][0].t.Before(cutoff) {
			s.slotSpeedSamples[slot] = s.slotSpeedSamples[slot][1:]
		}
	}
}

func (s *SyncStats) ClearSlotProgress(slot int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if slot >= 0 && slot < s.maxConcurrent {
		s.bytesInProgress[slot] = 0
	}
}

// GetSlotSpeed returns the download speed for a slot over the last 10 seconds.
func (s *SyncStats) GetSlotSpeed(slot int) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.getSlotSpeedLocked(slot)
}

func (s *SyncStats) getSlotSpeedLocked(slot int) int64 {
	if slot < 0 || slot >= s.maxConcurrent {
		return 0
	}
	samples := s.slotSpeedSamples[slot]
	if len(samples) < 2 {
		return 0
	}
	newest := samples[len(samples)-1]
	cutoff := newest.t.Add(-10 * time.Second)
	oldest := samples[0]
	for _, sample := range samples {
		if sample.t.After(cutoff) {
			break
		}
		oldest = sample
	}
	dt := newest.t.Sub(oldest.t).Seconds()
	if dt <= 0 {
		return 0
	}
	if speed := int64(float64(newest.bytes-oldest.bytes) / dt); speed > 0 {
		return speed
	}
	return 0
}

func (s *SyncStats) getGlobalSpeedLocked() int64 {
	if len(s.globalSpeedSamples) < 2 {
		return 0
	}
	newest := s.globalSpeedSamples[len(s.globalSpeedSamples)-1]
	cutoff := newest.t.Add(-10 * time.Second)
	oldest := s.globalSpeedSamples[0]
	for _, sample := range s.globalSpeedSamples {
		if sample.t.After(cutoff) {
			break
		}
		oldest = sample
	}
	dt := newest.t.Sub(oldest.t).Seconds()
	if dt <= 0 {
		return 0
	}
	if speed := int64(float64(newest.bytes-oldest.bytes) / dt); speed > 0 {
		return speed
	}
	return 0
}

func (s *SyncStats) getTotalBytesTransferred() int64 {
	// Must be called with lock held
	total := s.bytesDownloaded
	for i := range s.maxConcurrent {
		total += s.bytesInProgress[i]
	}
	return total
}

func (s *SyncStats) SetActivity(slot int, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if slot >= 0 && slot < s.maxConcurrent {
		s.activities[slot] = message
	}
}

func (s *SyncStats) ClearActivity(slot int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if slot >= 0 && slot < s.maxConcurrent {
		s.activities[slot] = ""
	}
}

func (s *SyncStats) Print() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Record global speed sample (download bytes only, not skipped)
	now := time.Now()
	globalBytes := s.bytesActuallyDownloaded
	for i := range s.maxConcurrent {
		globalBytes += s.bytesInProgress[i]
	}
	s.globalSpeedSamples = append(s.globalSpeedSamples, speedSample{t: now, bytes: globalBytes})
	cutoff := now.Add(-12 * time.Second)
	for len(s.globalSpeedSamples) > 1 && s.globalSpeedSamples[0].t.Before(cutoff) {
		s.globalSpeedSamples = s.globalSpeedSamples[1:]
	}

	// Count how many active activity lines we have
	activeCount := 0
	for i := range s.activeSlots {
		if s.activities[i] != "" {
			activeCount++
		}
	}

	// Calculate lines to print this time (activity lines + 4 stats lines)
	linesToPrint := activeCount + 4

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
	speed := s.getGlobalSpeedLocked()

	// Calculate percentage if total is known
	progressStr := ""
	if s.totalBytes > 0 {
		percentage := float64(totalTransferred) / float64(s.totalBytes) * 100
		progressStr = fmt.Sprintf(" (%.1f%%)", percentage)
	}

	// Print stats on separate lines
	fmt.Printf("%sFiles:%s %d / %d\n",
		colorBold, colorReset,
		s.filesChecked, s.filesTotal)
	fmt.Printf("       %s%d downloaded%s  %d skipped  %s%d deleted%s  %s%d errors%s\n",
		colorGreen,
		s.filesDownloaded,
		colorReset,
		s.filesSkipped,
		colorYellow,
		s.filesDeleted,
		colorReset,
		colorRed,
		s.filesErrors,
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
