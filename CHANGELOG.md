# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2025-02-04

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
