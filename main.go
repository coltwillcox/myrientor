package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	configFile    = "myrient-devices.json"
	maxConcurrent = 2 // equivalent to --transfers 2

	// ANSI color codes
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorWhite   = "\033[37m"
	colorBold    = "\033[1m"
	colorDim     = "\033[2m"
)

type Device struct {
	RemotePath string `json:"remote_path"`
	Sync       bool   `json:"sync"`
	LocalPath  string `json:"local_path"`
}

type Config struct {
	BaseURL string   `json:"base_url"`
	Devices []Device `json:"devices"`
}

type FileInfo struct {
	Name string
	Size int64
}

type SyncStats struct {
	mu               sync.Mutex
	filesChecked     int
	filesDownloaded  int
	filesDeleted     int
	filesSkipped     int
	bytesDownloaded  int64
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

	// Calculate stats
	elapsed := time.Since(s.startTime)
	speed := int64(0)
	if elapsed.Seconds() > 0 {
		speed = int64(float64(s.bytesDownloaded) / elapsed.Seconds())
	}

	// Calculate percentage if total is known
	progressStr := ""
	if s.totalBytes > 0 {
		percentage := float64(s.bytesDownloaded) / float64(s.totalBytes) * 100
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
		formatBytes(s.bytesDownloaded),
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

func formatBytesIfKnown(bytes int64) string {
	if bytes <= 0 {
		return "?"
	}
	return formatBytes(bytes)
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh%02dm%02ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%02ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
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

func main() {
	config, err := readConfigFile(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗ Error reading config file: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

	totalDevices := 0
	for _, device := range config.Devices {
		if device.Sync {
			totalDevices++
		}
	}

	fmt.Printf("%s%sStarting sync of %d device(s) from %s%s\n", colorBold, colorCyan, totalDevices, config.BaseURL, colorReset)
	fmt.Printf("%s═══════════════════════════════════════════════════════════════════════%s\n", colorDim, colorReset)

	currentDevice := 0
	for _, device := range config.Devices {
		if device.Sync {
			currentDevice++
			fmt.Printf("\n%s[%d/%d]%s %sSyncing: %s%s\n", colorBold, currentDevice, totalDevices, colorReset, colorMagenta, device.RemotePath, colorReset)
			fmt.Printf("%s───────────────────────────────────────────────────────────────────────%s\n", colorDim, colorReset)

			if err := syncDirectory(device, config.BaseURL); err != nil {
				fmt.Fprintf(os.Stderr, "%s✗ Error syncing %s: %v%s\n", colorRed, device.RemotePath, err, colorReset)
			}
		}
	}

	fmt.Printf("\n%s═══════════════════════════════════════════════════════════════════════%s\n", colorDim, colorReset)
	fmt.Printf("%s✓ Sync(s) completed%s\n", colorGreen, colorReset)
}

func readConfigFile(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func syncDirectory(device Device, baseURL string) error {
	stats := &SyncStats{startTime: time.Now()}

	// Create HTTP client with TLS verification disabled (--no-check-certificate)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 30 * time.Second,
	}

	// Get directory listing
	remoteURL := baseURL + device.RemotePath
	filesInfo, err := getDirectoryListing(client, remoteURL)
	if err != nil {
		return fmt.Errorf("failed to get directory listing: %w", err)
	}

	// Create local directory
	if err := os.MkdirAll(device.LocalPath, 0755); err != nil {
		return fmt.Errorf("failed to create local directory: %w", err)
	}

	// Filter files (exclude systeminfo.txt and directories)
	var filesToSync []FileInfo
	remoteFileSet := make(map[string]bool)
	totalSize := int64(0)

	for _, fileInfo := range filesInfo {
		if fileInfo.Name == "systeminfo.txt" || strings.HasSuffix(fileInfo.Name, "/") {
			continue
		}
		filesToSync = append(filesToSync, fileInfo)
		remoteFileSet[fileInfo.Name] = true

		// Pre-calculate total size for files that need downloading
		localFile := filepath.Join(device.LocalPath, fileInfo.Name)

		// Quick check if file needs downloading
		if needsSync(localFile, fileInfo.Size) {
			totalSize += fileInfo.Size
		}
	}

	// Set total bytes for progress tracking
	stats.SetTotalBytes(totalSize)

	// Set active slots to minimum of maxConcurrent and file count
	stats.activeSlots = min(maxConcurrent, len(filesToSync))
	if stats.activeSlots == 0 {
		stats.activeSlots = 1 // At least 1 slot for stats display
	}

	// Clean up obsolete local files
	if err := cleanupObsoleteFiles(device.LocalPath, remoteFileSet, stats); err != nil {
		fmt.Fprintf(os.Stderr, "%sWarning: failed to cleanup obsolete files: %v%s\n", colorYellow, err, colorReset)
	}

	// Sync files with concurrency control
	sem := make(chan struct{}, maxConcurrent)
	slotChan := make(chan int, maxConcurrent)

	// Initialize slot pool
	for i := range maxConcurrent {
		slotChan <- i
	}

	var wg sync.WaitGroup

	// Start stats printer
	stopStats := make(chan struct{})

	// Print initial empty lines for activities and stats
	initialLines := stats.activeSlots + 3
	for range initialLines {
		fmt.Println()
	}
	stats.lastPrintedLines = initialLines

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				stats.Print()
			case <-stopStats:
				return
			}
		}
	}()

	for _, fileInfo := range filesToSync {
		wg.Add(1)
		sem <- struct{}{}  // Acquire semaphore
		slot := <-slotChan // Get available slot

		go func(file FileInfo, activitySlot int) {
			defer wg.Done()
			defer func() {
				<-sem // Release semaphore
				stats.ClearActivity(activitySlot)
				slotChan <- activitySlot // Return slot
			}()

			stats.IncrementChecked()

			// URL-encode the filename for the remote request
			escapedFilename := url.PathEscape(file.Name)
			remoteFile := remoteURL + escapedFilename
			localFile := filepath.Join(device.LocalPath, file.Name)

			// Check if file needs downloading
			stats.SetActivity(activitySlot, fmt.Sprintf("%s→ Checking:%s %s", colorBlue, colorReset, file.Name))
			needsDownload, _, err := shouldDownload(client, remoteFile, localFile)
			if err != nil {
				stats.ClearActivity(activitySlot)
				fmt.Fprintf(os.Stderr, "\n%s✗ Error checking %s: %v%s\n", colorRed, file.Name, err, colorReset)
				// Re-print activity lines after error
				for i := 0; i < maxConcurrent; i++ {
					fmt.Println()
				}
				return
			}

			if needsDownload {
				stats.SetActivity(activitySlot, fmt.Sprintf("%s↓ Downloading:%s %s", colorCyan, colorReset, file.Name))
				bytes, err := downloadFile(client, remoteFile, localFile)
				if err != nil {
					stats.ClearActivity(activitySlot)
					fmt.Fprintf(os.Stderr, "\n%s✗ Error downloading %s: %v%s\n", colorRed, file.Name, err, colorReset)
					// Re-print activity lines after error
					for i := 0; i < maxConcurrent; i++ {
						fmt.Println()
					}
					return
				}
				stats.IncrementDownloaded(bytes)
				stats.SetActivity(activitySlot, fmt.Sprintf("%s✓ Downloaded:%s %s %s(%s)%s", colorGreen, colorReset, file.Name, colorDim, formatBytes(bytes), colorReset))
			} else {
				stats.IncrementSkipped()
				stats.ClearActivity(activitySlot)
			}
		}(fileInfo, slot)
	}

	wg.Wait()
	close(stopStats)

	// Print final stats
	stats.Print()
	fmt.Println()

	fmt.Printf("\n%s✓ Sync complete%s\n", colorGreen, colorReset)

	return nil
}

func cleanupObsoleteFiles(localPath string, remoteFiles map[string]bool, stats *SyncStats) error {
	// Read local directory
	entries, err := os.ReadDir(localPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Directory doesn't exist yet, nothing to clean
		}
		return err
	}

	deletedCount := 0

	// Check each local file
	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip subdirectories
		}

		filename := entry.Name()

		// Skip ignored files
		if filename == "systeminfo.txt" {
			continue
		}

		// If file doesn't exist remotely, delete it
		if !remoteFiles[filename] {
			localFile := filepath.Join(localPath, filename)
			if err := os.Remove(localFile); err != nil {
				fmt.Fprintf(os.Stderr, "%s✗ Failed to remove: %s (%v)%s\n", colorRed, filename, err, colorReset)
			} else {
				stats.IncrementDeleted()
				deletedCount++
			}
		}
	}

	if deletedCount > 0 {
		fmt.Printf("%s✓ Cleaned up %d obsolete file(s)%s\n", colorYellow, deletedCount, colorReset)
	}

	return nil
}

func getDirectoryListing(client *http.Client, dirURL string) ([]FileInfo, error) {
	resp, err := client.Get(dirURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse simple directory listing (HTML)
	var files []FileInfo
	lines := strings.Split(string(body), "\n")

	for i := range lines {
		line := lines[i]

		// Look for href="..." patterns
		if strings.Contains(line, "href=") {
			start := strings.Index(line, "href=\"")
			if start == -1 {
				continue
			}
			start += 6
			end := strings.Index(line[start:], "\"")
			if end == -1 {
				continue
			}

			href := line[start : start+end]

			// Skip parent directory, absolute URLs, anchors, and query strings
			if href == "../" ||
				strings.HasPrefix(href, "http") ||
				strings.HasPrefix(href, "#") ||
				strings.HasPrefix(href, "/") ||
				strings.HasPrefix(href, "?") {
				continue
			}

			// Skip if href contains query string (e.g., "file.zip?param=value")
			if strings.Contains(href, "?") {
				continue
			}

			// Unescape URL encoding (e.g., %5B -> [, %20 -> space)
			unescaped, err := url.QueryUnescape(href)
			if err != nil {
				// If unescaping fails, use the original
				unescaped = href
			}

			// Look for size in the next line with class="size"
			size := int64(0)
			for j := i; j < len(lines) && j < i+3; j++ {
				if strings.Contains(lines[j], "class=\"size\"") {
					sizeStr := extractSizeFromHTML(lines[j])
					size = parseSizeString(sizeStr)
					break
				}
			}

			files = append(files, FileInfo{
				Name: unescaped,
				Size: size,
			})
		}
	}

	return files, nil
}

func extractSizeFromHTML(line string) string {
	// Extract content between <td class="size"> and </td>
	start := strings.Index(line, "<td class=\"size\">")
	if start == -1 {
		return ""
	}
	start += len("<td class=\"size\">")

	end := strings.Index(line[start:], "</td>")
	if end == -1 {
		return ""
	}

	return strings.TrimSpace(line[start : start+end])
}

func parseSizeString(sizeStr string) int64 {
	if sizeStr == "" || sizeStr == "-" {
		return 0
	}

	// Parse sizes like "10.3 KiB", "735 B", "1.5 MiB", "2.1 GiB"
	parts := strings.Fields(sizeStr)
	if len(parts) != 2 {
		return 0
	}

	var value float64
	fmt.Sscanf(parts[0], "%f", &value)

	unit := parts[1]
	multiplier := int64(1)

	switch unit {
	case "B":
		multiplier = 1
	case "KiB":
		multiplier = 1024
	case "MiB":
		multiplier = 1024 * 1024
	case "GiB":
		multiplier = 1024 * 1024 * 1024
	case "TiB":
		multiplier = 1024 * 1024 * 1024 * 1024
	}

	return int64(value * float64(multiplier))
}

func needsSync(localPath string, expectedSize int64) bool {
	info, err := os.Stat(localPath)
	if os.IsNotExist(err) {
		return true // File doesn't exist
	}
	if err != nil {
		return true // Error reading file, better to re-download
	}

	// If sizes don't match, needs sync
	if expectedSize > 0 && info.Size() != expectedSize {
		return true
	}

	return false // File exists and size matches
}

func shouldDownload(client *http.Client, remoteURL, localPath string) (bool, int64, error) {
	// Check if local file exists
	localInfo, err := os.Stat(localPath)
	if os.IsNotExist(err) {
		// File doesn't exist, get remote size
		resp, err := client.Head(remoteURL)
		if err != nil {
			return false, 0, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return false, 0, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
		}

		return true, resp.ContentLength, nil
	}
	if err != nil {
		return false, 0, err
	}

	// Get remote file info
	resp, err := client.Head(remoteURL)
	if err != nil {
		return false, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, 0, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Compare sizes
	remoteSize := resp.ContentLength
	localSize := localInfo.Size()

	if remoteSize != localSize {
		return true, remoteSize, nil
	}

	// Compare modification times if available
	if lastModified := resp.Header.Get("Last-Modified"); lastModified != "" {
		remoteTime, err := http.ParseTime(lastModified)
		if err == nil {
			if remoteTime.After(localInfo.ModTime()) {
				return true, remoteSize, nil
			}
		}
	}

	return false, 0, nil // File is up to date
}

func downloadFile(client *http.Client, fileURL, filepath string) (int64, error) {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return 0, err
	}
	defer out.Close()

	// Download the file
	resp, err := client.Get(fileURL)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Copy the content
	bytes, err := io.Copy(out, resp.Body)
	if err != nil {
		return 0, err
	}

	// Set modification time if available
	if lastModified := resp.Header.Get("Last-Modified"); lastModified != "" {
		if modTime, err := http.ParseTime(lastModified); err == nil {
			os.Chtimes(filepath, modTime, modTime)
		}
	}

	return bytes, nil
}
