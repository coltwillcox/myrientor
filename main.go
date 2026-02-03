package main

import (
	"flag"
	"fmt"
	"os"
)

// Version info - injected at build time via ldflags
var version = "dev"

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
	colorBold    = "\033[1m"
	colorDim     = "\033[2m"
)

func main() {
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("myrientor %s\n", version)
		os.Exit(0)
	}

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
