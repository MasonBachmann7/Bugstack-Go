# Contributing to bugstack-go

Thank you for your interest in contributing!

## Development Setup

```bash
git clone https://github.com/MasonBachmann7/bugstack-go.git
cd bugstack-go
go mod download
```

## Running Tests

```bash
go test ./...
```

## Code Style

Follow standard Go conventions. Run `go vet` and `gofmt` before submitting.

## Pull Requests

1. Fork the repo and create your branch from `main`
2. Add tests for any new functionality
3. Ensure all tests pass (`go test ./...`)
4. Run `go vet ./...`
5. Submit your PR

## Reporting Issues

Use [GitHub Issues](https://github.com/MasonBachmann7/bugstack-go/issues).
