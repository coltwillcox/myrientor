package main

import (
	"fmt"
	"time"
)

func formatBytesIfKnown(bytes int64) string {
	if bytes <= 0 {
		return "?"
	}
	return formatBytes(bytes)
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if days > 0 {
		return fmt.Sprintf("%dd %02dh %02dm %02ds", days, h, m, s)
	}
	if h > 0 {
		return fmt.Sprintf("%dh %02dm %02ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %02ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

// fitInTerminal crops name so that (overhead + len(name)) fits within the
// terminal width. overhead is the number of display columns taken by the
// fixed parts of the line (prefix, separator, suffix). If cropping is
// necessary, the last character is replaced with "…".
func fitInTerminal(name string, overhead int) string {
	maxLen := terminalWidth() - overhead
	if maxLen < 1 {
		maxLen = 1
	}
	runes := []rune(name)
	if len(runes) <= maxLen {
		return name
	}
	if maxLen == 1 {
		return "…"
	}
	return string(runes[:maxLen-1]) + "…"
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
