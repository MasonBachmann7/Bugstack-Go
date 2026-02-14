# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability, please report it responsibly:

**Email:** security@bugstack.dev

Please do **not** open a public GitHub issue for security vulnerabilities.

We will respond within 48 hours.

## Supported Versions

| Version | Supported |
|---------|-----------|
| 1.x     | Yes       |

## Security Design

- Zero external dependencies — only Go stdlib
- No cookies, IP addresses, headers, or user data captured by default
- `BeforeSend` hook lets you filter every event before it leaves your app
- `DryRun` mode for full transparency
- The SDK never panics — all panics are recovered internally
- All data transmission uses HTTPS
