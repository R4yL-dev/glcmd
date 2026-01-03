# Changelog

All notable changes to glcmd are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [0.2.0] - 2026-01-03

### Added
- Unified HTTP API server on port 8080 with 6 endpoints
- Comprehensive test suite for critical components
- Automatic sensor change detection with transaction safety
- Structured logging with slog
- Database health checks
- PostgreSQL support (configuration ready)
- Graceful shutdown with signal handling
- Circuit breaker pattern for API error recovery
- Exponential backoff retry logic

### Changed
- **BREAKING**: Application now runs exclusively in daemon mode
- **BREAKING**: Unified API servers to single port 8080
- **BREAKING**: `GLCMD_HEALTHCHECK_PORT` → `GLCMD_API_PORT`
- Refactored architecture with improved error handling
- Enhanced database configuration

### Fixed
- Database connection retry logic
- Transaction atomicity for sensor changes
- Memory cleanup on shutdown

### Removed
- Separate health check and metrics servers
- Standalone/one-time query mode (daemon mode only)

## [0.1.1] - 2025-05-04

Initial release with basic functionality.
