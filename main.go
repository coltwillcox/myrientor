package main

import (
	"flag"
	"fmt"
	"os"
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

	totalDevices := remoteConfig.SyncableCount()

	fmt.Printf("%s%sStarting sync of %d device(s) from %s%s\n", colorBold, colorCyan, totalDevices, remoteConfig.BaseURL, colorReset)
	fmt.Printf("%s═══════════════════════════════════════════════════════════════════════%s\n", colorDim, colorReset)

	currentDevice := 0
	for _, device := range remoteConfig.Devices {
		if device.ShouldSync() {
			currentDevice++
			fmt.Printf("\n%s[%d/%d]%s %sSyncing: %s%s\n", colorBold, currentDevice, totalDevices, colorReset, colorMagenta, device.RemotePath, colorReset)
			fmt.Printf("%s───────────────────────────────────────────────────────────────────────%s\n", colorDim, colorReset)

			if err := syncDirectory(device, remoteConfig.BaseURL, maxConcurrent, errLog); err != nil {
				errLog.Log("Error syncing %s: %v", device.RemotePath, err)
			}
		}
	}

	fmt.Printf("\n%s═══════════════════════════════════════════════════════════════════════%s\n", colorDim, colorReset)

	// Display error summary
	errorCount := errLog.Count()
	if errorCount > 0 {
		fmt.Printf("%s✓ Sync(s) completed with %d error(s)%s\n", colorYellow, errorCount, colorReset)
		fmt.Printf("%s  See: %s%s\n", colorDim, errLog.Filename(), colorReset)
	} else {
		fmt.Printf("%s✓ Sync(s) completed%s\n", colorGreen, colorReset)
	}
}
