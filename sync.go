package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	downloadMaxRetries   = 3
	downloadStallTimeout = 30 * time.Second
)

type FileInfo struct {
	Name string
	Size int64
}

func syncDirectory(device Device, baseURL string, maxConcurrent int, errLog *ErrorLogger) error {
	stats := NewSyncStats(maxConcurrent)

	// Client for quick operations (HEAD requests, directory listings)
	// with TLS verification disabled (--no-check-certificate)
	quickClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 30 * time.Second,
	}

	// Client for downloads - connection timeouts but no overall timeout for large files
	// with TLS verification disabled (--no-check-certificate)
	downloadClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 10 * time.Second,
			IdleConnTimeout:     90 * time.Second,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: maxConcurrent,
		},
		Timeout: 0,
	}

	// Get directory listing
	remoteURL := baseURL + device.RemotePath
	filesInfo, err := getDirectoryListing(quickClient, remoteURL)
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

		totalSize += fileInfo.Size
	}

	// Set total bytes for progress tracking
	stats.SetTotalBytes(totalSize)

	// Set total file count and active slots
	stats.SetFilesTotal(len(filesToSync))
	stats.activeSlots = min(maxConcurrent, len(filesToSync))
	if stats.activeSlots == 0 {
		stats.activeSlots = 1 // At least 1 slot for stats display
	}

	// Clean up obsolete local files
	if err := cleanupObsoleteFiles(device.LocalPath, remoteFileSet, stats, errLog); err != nil {
		errLog.Log("Warning: failed to cleanup obsolete files in %s: %v", device.LocalPath, err)
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
	initialLines := stats.activeSlots + 6
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
			needsDownload, err := shouldDownload(quickClient, remoteFile, localFile)
			if err != nil {
				stats.IncrementErrors()
				stats.ClearActivity(activitySlot)
				errLog.Log("Error checking %s: %v", file.Name, err)
				return
			}

			if needsDownload {
				// Progress callback for this file
				onProgress := func(written, total int64) {
					stats.SetSlotProgress(activitySlot, written)
					speed := stats.GetSlotSpeed(activitySlot)
					if total > 0 {
						pct := float64(written) / float64(total) * 100
						stats.SetActivity(activitySlot, fmt.Sprintf("%s↓%s %s %s%.0f%% %s/%s @ %s/s%s",
							colorCyan, colorReset,
							file.Name,
							colorDim, pct,
							formatBytes(written), formatBytes(total),
							formatBytes(speed),
							colorReset))
					} else {
						stats.SetActivity(activitySlot, fmt.Sprintf("%s↓%s %s %s%s @ %s/s%s",
							colorCyan, colorReset,
							file.Name,
							colorDim, formatBytes(written),
							formatBytes(speed),
							colorReset))
					}
				}

				bytes, err := downloadFile(downloadClient, remoteFile, localFile, onProgress)
				stats.ClearSlotProgress(activitySlot) // Clear in-progress bytes when done
				if err != nil {
					stats.IncrementErrors()
					stats.ClearActivity(activitySlot)
					errLog.Log("Error downloading %s: %v", file.Name, err)
					return
				}
				stats.IncrementDownloaded(activitySlot, bytes)
				stats.SetActivity(activitySlot, fmt.Sprintf("%s✓%s %s %s(%s)%s", colorGreen, colorReset, file.Name, colorDim, formatBytes(bytes), colorReset))
			} else {
				stats.IncrementSkipped(file.Size)
				stats.ClearActivity(activitySlot)
			}
		}(fileInfo, slot)
	}

	wg.Wait()
	close(stopStats)

	// Print final stats
	stats.Print()
	fmt.Printf("\n%s✓ Sync complete%s\n", colorGreen, colorReset)
	fmt.Println()

	return nil
}

func cleanupObsoleteFiles(localPath string, remoteFiles map[string]bool, stats *SyncStats, errLog *ErrorLogger) error {
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
				stats.IncrementErrors()
				errLog.Log("Failed to remove %s: %v", localFile, err)
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

func shouldDownload(client *http.Client, remoteURL, localPath string) (bool, error) {
	// Check if local file exists
	localInfo, err := os.Stat(localPath)
	if os.IsNotExist(err) {
		// File doesn't exist, get remote size
		resp, err := client.Head(remoteURL)
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return false, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
		}

		return true, nil
	}
	if err != nil {
		return false, err
	}

	// Get remote file info
	resp, err := client.Head(remoteURL)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Compare sizes
	remoteSize := resp.ContentLength
	localSize := localInfo.Size()

	if remoteSize != localSize {
		return true, nil
	}

	// Compare modification times if available
	if lastModified := resp.Header.Get("Last-Modified"); lastModified != "" {
		remoteTime, err := http.ParseTime(lastModified)
		if err == nil {
			if remoteTime.After(localInfo.ModTime()) {
				return true, nil
			}
		}
	}

	return false, nil // File is up to date
}

// downloadFile downloads a file with automatic retry on stall or transient error.
// It uses HTTP Range requests to resume from where a failed attempt left off.
// Returns total bytes written to the file.
func downloadFile(client *http.Client, fileURL, filePath string, onProgress func(written, total int64)) (int64, error) {
	var totalInFile int64
	for attempt := 0; attempt <= downloadMaxRetries; attempt++ {
		n, err := downloadAttempt(client, fileURL, filePath, totalInFile, onProgress)
		totalInFile = n
		if err == nil {
			return totalInFile, nil
		}
		if attempt == downloadMaxRetries {
			return totalInFile, err
		}
	}
	return totalInFile, nil
}

// downloadAttempt performs a single download attempt starting at offset.
// If the server supports Range requests and offset > 0, it resumes from offset;
// otherwise it restarts from the beginning.
// Returns total bytes present in the file after this attempt.
func downloadAttempt(client *http.Client, fileURL, filePath string, offset int64, onProgress func(written, total int64)) (int64, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		return offset, err
	}
	if offset > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", offset))
	}

	resp, err := client.Do(req)
	if err != nil {
		return offset, err
	}
	defer resp.Body.Close()

	var (
		out        *os.File
		fileOffset int64 // actual byte offset we write from in the file
		totalSize  int64 // total file size (for progress reporting)
	)

	switch resp.StatusCode {
	case http.StatusPartialContent:
		// Server honours the Range request; resume writing from offset
		totalSize = parseTotalFromContentRange(resp.Header.Get("Content-Range"))
		fileOffset = offset
		out, err = os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return offset, err
		}
		if _, err = out.Seek(fileOffset, io.SeekStart); err != nil {
			out.Close()
			return offset, err
		}
	case http.StatusOK:
		// Server does not support Range; restart from the beginning
		fileOffset = 0
		totalSize = resp.ContentLength
		out, err = os.Create(filePath)
		if err != nil {
			return 0, err
		}
	default:
		return offset, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}
	defer out.Close()

	// Stall watchdog: cancel the context if no data arrives for stallTimeout
	var (
		lastReadMu sync.Mutex
		lastRead   = time.Now()
	)
	watchdogDone := make(chan struct{})
	defer close(watchdogDone)
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-watchdogDone:
				return
			case <-ticker.C:
				lastReadMu.Lock()
				stalled := time.Since(lastRead) > downloadStallTimeout
				lastReadMu.Unlock()
				if stalled {
					cancel()
				}
			}
		}
	}()

	// Read loop with stall tracking and progress reporting
	written := int64(0)
	buf := make([]byte, 32*1024)
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			lastReadMu.Lock()
			lastRead = time.Now()
			lastReadMu.Unlock()
			if _, werr := out.Write(buf[:n]); werr != nil {
				return fileOffset + written, werr
			}
			written += int64(n)
			if onProgress != nil {
				onProgress(fileOffset+written, totalSize)
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return fileOffset + written, rerr
		}
	}

	// Set modification time if available
	if lastModified := resp.Header.Get("Last-Modified"); lastModified != "" {
		if modTime, err := http.ParseTime(lastModified); err == nil {
			os.Chtimes(filePath, modTime, modTime)
		}
	}

	return fileOffset + written, nil
}

// parseTotalFromContentRange extracts the total file size from a Content-Range header.
// Format: "bytes X-Y/Z" → returns Z.
func parseTotalFromContentRange(contentRange string) int64 {
	slash := strings.LastIndex(contentRange, "/")
	if slash < 0 {
		return 0
	}
	var total int64
	fmt.Sscanf(contentRange[slash+1:], "%d", &total)
	return total
}
