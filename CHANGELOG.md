# Changelog

All notable changes to this project will be documented in this file.

## [1.0.0] - 2026-02-13

### Added

- Core SDK with `bugstack.Init()`, `bugstack.CaptureError()`, `bugstack.CaptureMessage()`
- Panic recovery with `bugstack.Recover()`
- Background goroutine transport with buffered channel and retry with exponential backoff
- SHA-256 error fingerprinting and client-side deduplication
- `BeforeSend` hook for event inspection/modification/filtering
- `IgnoredErrors` for skipping errors by message substring
- `DryRun` mode for transparent debugging
- `Enabled` kill switch
- net/http middleware with panic recovery
- Gin middleware adapter
- Echo middleware adapter
- Zero external dependencies
