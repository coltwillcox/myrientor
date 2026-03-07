package main

import (
	"fmt"
	"strings"
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

// activityLine builds a full-terminal-width activity line.
// prefix is the already-rendered left label (with ANSI codes); prefixCols is
// its visible column count. name is placed after the prefix. suffix is placed
// flush to the right edge (ASCII only). The gap is filled with dots.
// If suffix is empty the dots fill to the terminal edge.
// If name is too long it is cropped with "…" to leave room for the dots.
func activityLine(prefix string, prefixCols int, name, suffix string) string {
	tw := terminalWidth()
	nameCols := len([]rune(name))
	suffixCols := len(suffix)
	dots := tw - prefixCols - nameCols - suffixCols
	if dots < 1 {
		excess := 1 - dots
		runes := []rune(name)
		if excess+1 >= len(runes) {
			name = "…"
			nameCols = 1
		} else {
			name = string(runes[:len(runes)-excess-1]) + "…"
			nameCols = len(runes) - excess
		}
		_ = nameCols
		dots = 1
	}
	if suffix == "" {
		return prefix + name + colorDim + strings.Repeat(".", dots) + colorReset
	}
	return prefix + name + colorDim + strings.Repeat(".", dots) + suffix + colorReset
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

func devicePanel(index, total int, path string) string {
	tw := terminalWidth()
	top := panelTopLabeled(fmt.Sprintf("%d/%d", index, total))

	// │  Syncing: <path>   │
	const syncPrefix = "  Syncing: "
	inner := tw - 2
	maxPathCols := max(inner-len(syncPrefix), 1)
	runes := []rune(path)
	if len(runes) > maxPathCols {
		if maxPathCols == 1 {
			path = "…"
		} else {
			path = string(runes[:maxPathCols-1]) + "…"
		}
		runes = []rune(path)
	}
	padding := max(inner-len(syncPrefix)-len(runes), 0)
	middle := colorCyan + "│" + colorReset + syncPrefix + colorMagenta + path + colorReset + strings.Repeat(" ", padding) + colorCyan + "│" + colorReset

	// └──...──┘
	bottom := colorCyan + "└" + strings.Repeat("─", tw-2) + "┘" + colorReset

	return top + "\n" + middle + "\n" + bottom
}

// stripANSI removes ANSI escape sequences from s so its display width can be measured.
func stripANSI(s string) string {
	var b strings.Builder
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '\033' && i+1 < len(runes) && runes[i+1] == '[' {
			i += 2
			for i < len(runes) && runes[i] != 'm' {
				i++
			}
		} else {
			b.WriteRune(runes[i])
		}
	}
	return b.String()
}

func panelTopLabeled(label string) string {
	tw := terminalWidth()
	labelCols := len([]rune(label))
	dashes := max(tw-8-labelCols, 0)
	return colorCyan + "┌──[" + colorBold + " " + label + " " + colorReset + colorCyan + "]" + strings.Repeat("─", dashes) + "┐" + colorReset
}

func panelTop() string {
	tw := terminalWidth()
	return colorCyan + "┌" + strings.Repeat("─", tw-2) + "┐" + colorReset
}

// panelLine wraps content in a panel row: │  content...padding │
// Content may contain ANSI codes; display width is measured by stripping them.
func panelLine(content string) string {
	tw := terminalWidth()
	displayCols := len([]rune(stripANSI(content)))
	available := tw - 4 // │ + 2-space left margin + │
	padding := max(available-displayCols, 0)
	return colorCyan + "│" + colorReset + "  " + content + strings.Repeat(" ", padding) + colorCyan + "│" + colorReset
}

func panelBottom() string {
	tw := terminalWidth()
	return colorCyan + "└" + strings.Repeat("─", tw-2) + "┘" + colorReset
}

func separatorDouble() string {
	return colorDim + strings.Repeat("═", terminalWidth()) + colorReset
}

func separatorSingle() string {
	return colorDim + strings.Repeat("─", terminalWidth()) + colorReset
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
