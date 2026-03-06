# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.12.0] - 2026-03-06

### Changed
- Device title rendered in a single-line box panel: `┌──[ N/M ]──...──┐ │  Syncing: path │ └──...──┘`
- Stats footer (Files, Transfer, Time) rendered in a matching panel, separated from slot activity lines by a blank line
- Panel borders fill the full terminal width dynamically

## [0.11.1] - 2026-03-06

### Changed
- Separator lines (`═══` and `───`) now fill the full terminal width dynamically

## [0.11.0] - 2026-03-06

### Added
- Recursive directory sync: subdirectories within a remote path are now traversed and synced automatically
- Scanning progress shown in-place while the remote directory tree is being crawled

### Changed
- `-sync` flag now matches by either `local_path` or `remote_path`; all matching devices are synced
- Redundant HEAD request removed when local file is missing (file is known to exist from the directory listing)
- Stats display refresh rate increased to 100 ms

## [0.10.0] - 2026-03-06

### Changed
- Files are now downloaded to `local_path/remote_path/` instead of `local_path/`, mirroring the remote directory structure locally
- `-sync` flag now matches devices by `remote_path` instead of `local_path`
- Error log entries now include both `local_path` and `remote_path` for easier triage

### Removed
- `FindByLocalPath` — replaced by `FindByRemotePath`

## [0.8.3] - 2026-03-05

### Changed
- Updated remote.json paths to match current Myrient directory structure

## [0.8.2] - 2026-03-05

### Changed
- Updated remote.json paths to match current Myrient directory structure

## [0.8.1] - 2026-03-05

### Changed
- Error log entries now prefixed with `device.LocalPath` for easier triage

## [0.8.0] - 2026-03-04

### Added
- Drain hotkey (`q` / `Q`): stops queuing new files but lets active downloads finish
- When draining, remaining queued devices are also skipped
- `[ draining ]` indicator shown in stats display while drain is active
- Activity line length capped to terminal width; filename cropped with `…` if needed
- Native drain hotkey support on macOS (`TIOCGETA`/`TIOCSETA`) and Windows (console API)
- Terminal width queried via `TIOCGWINSZ` on Linux/macOS and `GetConsoleScreenBufferInfo` on Windows

## [0.7.0] - 2026-03-03

### Added
- Total file count shown in Files stat (`N / total`)
- ETA displayed below elapsed time, based on overall processing rate
- Automatic download retry (up to 3 attempts) with HTTP Range resume on stall or error
- Stall detection: cancels and retries if no data received for 30 seconds

### Changed
- Files, Transfer, and Time stat lines each split into two lines for readability
- `formatDuration` now includes days (e.g. `2d 03h 15m 42s`)

## [0.6.0] - 2026-03-03

### Changed
- Transfer speed (global and per-file) now uses a 10-second sliding window instead of total elapsed time
- Global transfer speed excludes already-checked (skipped) files

### Added
- Per-file download speed shown in each slot's activity line

## [0.5.0] - 2026-03-02

### Added
- Errors counter in stats display

## [0.4.0] - 2026-02-04

### Added
- `-sync` flag to sync specific device by `local_path`
- `FindByLocalPath()` method to find devices by local path

## [0.3.0] - 2026-02-04

### Changed
- Refactored config reading functions to use constants for filenames
- Added helper methods `ShouldSync()` and `SyncableCount()` to config types
- Changed remote.json values for local paths

## [0.2.1] - 2026-02-04

### Added
- Local configuration file (`local.json`) for user settings
- `-concurrent` flag to set maximum concurrent downloads
- Error logging to timestamped file instead of console output
- Real-time transfer speed calculation from active downloads
- Download progress display with percentage and size ratio
- GitHub Actions workflow for multi-platform releases
- Version flag (`-version`) with build-time injection

### Changed
- Split monolithic `main.go` into separate modules (`config.go`, `stats.go`, `sync.go`, `format.go`, `errorlog.go`)
- HTTP client configuration with separate clients for quick operations and large downloads
- Large file downloads no longer timeout (connection-level timeouts only)

### Fixed
- Stats display drifting when fewer files than maxConcurrent
- Empty activity lines when downloads complete
- HTTP timeout errors on large file downloads
