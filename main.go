package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Version info - injected at build time via ldflags
var version = "dev"

const (
	defaultMaxConcurrent = 2

	// ANSI color codes
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorBold    = "\033[1m"
	colorDim     = "\033[2m"
)

func main() {
	showVersion := flag.Bool("version", false, "Show version information")
	maxConcurrentFlag := flag.Int("concurrent", 0, "Maximum concurrent downloads")
	syncFlag := flag.String("sync", "", "Sync specific device by remote_path")
	flag.Parse()

	if *showVersion {
		fmt.Printf("myrientor %s\n", version)
		os.Exit(0)
	}

	// Determine maxConcurrent: flag > config file > default
	maxConcurrent := defaultMaxConcurrent
	if localConfig, err := readLocalConfigFile(); err == nil && localConfig.MaxConcurrent > 0 {
		maxConcurrent = localConfig.MaxConcurrent
	}
	if *maxConcurrentFlag > 0 {
		maxConcurrent = *maxConcurrentFlag
	}

	// Initialize error logger
	errLog := NewErrorLogger()
	defer errLog.Close()

	remoteConfig, err := readRemoteConfigFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗ Error reading config file: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

	// Build list of devices to sync
	var devicesToSync []Device
	if *syncFlag != "" {
		devicesToSync = remoteConfig.FindAllByPath(*syncFlag)
		if len(devicesToSync) == 0 {
			fmt.Fprintf(os.Stderr, "%s✗ No syncable device found matching: %s%s\n", colorRed, *syncFlag, colorReset)
			os.Exit(1)
		}
	} else {
		// Sync all enabled devices
		for _, device := range remoteConfig.Devices {
			if device.ShouldSync() {
				devicesToSync = append(devicesToSync, device)
			}
		}
	}

	totalDevices := len(devicesToSync)

	fmt.Printf("%s%sStarting sync of %d device(s) from %s%s\n", colorBold, colorCyan, totalDevices, remoteConfig.BaseURL, colorReset)
	fmt.Println(separatorDouble())

	overallStart := time.Now()
	var total SyncSummary
	devicesSynced := 0

	for i, device := range devicesToSync {
		fmt.Printf("\n%s\n", devicePanel(i+1, totalDevices, device.RemotePath))

		drained, summary, err := syncDirectory(device, remoteConfig.BaseURL, maxConcurrent, errLog)
		if err != nil {
			localDir := filepath.Join(device.LocalPath, device.RemotePath)
			errLog.Log("%s: error syncing: %v", localDir, err)
		}
		devicesSynced++
		total.FilesDownloaded += summary.FilesDownloaded
		total.FilesSkipped += summary.FilesSkipped
		total.FilesDeleted += summary.FilesDeleted
		total.FilesErrors += summary.FilesErrors
		total.BytesDownloaded += summary.BytesDownloaded
		total.BytesSkipped += summary.BytesSkipped

		fmt.Println(separatorSingle())
		if drained {
			break
		}
	}

	elapsed := time.Since(overallStart)
	totalBytes := total.BytesDownloaded + total.BytesSkipped

	fmt.Println()
	fmt.Println(panelTopLabeled("SUMMARY"))
	fmt.Println(panelLine(fmt.Sprintf("%sFiles:%s    %s%d downloaded%s  %d skipped  %s%d deleted%s  %s%d errors%s",
		colorBold, colorReset,
		colorGreen, total.FilesDownloaded, colorReset,
		total.FilesSkipped,
		colorYellow, total.FilesDeleted, colorReset,
		colorRed, total.FilesErrors, colorReset)))
	fmt.Println(panelLine(fmt.Sprintf("%sTransfer:%s %s%s downloaded%s  %s skipped  %s total",
		colorBold, colorReset,
		colorCyan, formatBytes(total.BytesDownloaded), colorReset,
		formatBytes(total.BytesSkipped),
		formatBytes(totalBytes))))
	fmt.Println(panelLine(fmt.Sprintf("%sTime:%s     %s%s%s",
		colorBold, colorReset, colorBlue, formatDuration(elapsed), colorReset)))
	fmt.Println(panelLine(fmt.Sprintf("%sDevices:%s  %d synced",
		colorBold, colorReset, devicesSynced)))
	fmt.Print(panelBottom())
	fmt.Println()
	fmt.Println()

	// Display error summary
	errorCount := errLog.Count()
	if errorCount > 0 {
		fmt.Printf("%s✓ Sync(s) completed with %d error(s)%s\n", colorYellow, errorCount, colorReset)
		fmt.Printf("%s  See: %s%s\n", colorDim, errLog.Filename(), colorReset)
	} else {
		fmt.Printf("%s✓ Sync(s) completed%s\n", colorGreen, colorReset)
	}
}
