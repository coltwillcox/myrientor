package main

import (
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

type FileInfo struct {
	Name string
	Size int64
}

// ProgressWriter wraps an io.Writer and reports progress via callback
type ProgressWriter struct {
	writer     io.Writer
	total      int64
	written    int64
	onProgress func(written, total int64)
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	pw.written += int64(n)
	if pw.onProgress != nil {
		pw.onProgress(pw.written, pw.total)
	}
	return n, err
}

func syncDirectory(device Device, baseURL string, errLog *ErrorLogger) error {
	stats := &SyncStats{startTime: time.Now()}

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
			needsDownload, err := shouldDownload(quickClient, remoteFile, localFile)
			if err != nil {
				stats.ClearActivity(activitySlot)
				errLog.Log("Error checking %s: %v", file.Name, err)
				return
			}

			if needsDownload {
				// Progress callback for this file
				onProgress := func(written, total int64) {
					stats.SetSlotProgress(activitySlot, written)
					if total > 0 {
						pct := float64(written) / float64(total) * 100
						stats.SetActivity(activitySlot, fmt.Sprintf("%s↓%s %s %s%.0f%% %s/%s%s",
							colorCyan, colorReset,
							file.Name,
							colorDim, pct,
							formatBytes(written), formatBytes(total),
							colorReset))
					} else {
						stats.SetActivity(activitySlot, fmt.Sprintf("%s↓%s %s %s%s%s",
							colorCyan, colorReset,
							file.Name,
							colorDim, formatBytes(written),
							colorReset))
					}
				}

				bytes, err := downloadFile(downloadClient, remoteFile, localFile, onProgress)
				stats.ClearSlotProgress(activitySlot) // Clear in-progress bytes when done
				if err != nil {
					stats.ClearActivity(activitySlot)
					errLog.Log("Error downloading %s: %v", file.Name, err)
					return
				}
				stats.IncrementDownloaded(bytes)
				stats.SetActivity(activitySlot, fmt.Sprintf("%s✓%s %s %s(%s)%s", colorGreen, colorReset, file.Name, colorDim, formatBytes(bytes), colorReset))
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

func downloadFile(client *http.Client, fileURL, filepath string, onProgress func(written, total int64)) (int64, error) {
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

	// Copy the content with progress tracking
	var written int64
	if onProgress != nil {
		pw := &ProgressWriter{
			writer:     out,
			total:      resp.ContentLength,
			onProgress: onProgress,
		}
		written, err = io.Copy(pw, resp.Body)
	} else {
		written, err = io.Copy(out, resp.Body)
	}
	if err != nil {
		return 0, err
	}

	// Set modification time if available
	if lastModified := resp.Header.Get("Last-Modified"); lastModified != "" {
		if modTime, err := http.ParseTime(lastModified); err == nil {
			os.Chtimes(filepath, modTime, modTime)
		}
	}

	return written, nil
}
