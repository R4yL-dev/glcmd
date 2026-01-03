# Changelog

All notable changes to glcmd are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [0.3.0] - 2026-01-03

### Added
- **API Versioning**: All data endpoints now use `/v1` prefix for API stability
- **CORS Middleware**: Cross-Origin Resource Sharing support for web frontends
- **Centralized Configuration**: New `internal/config` package for unified configuration management
- **Domain Constants**: Named constants for measurement colors, trend arrows, and glucose units
- **Enhanced Health Check**: Health endpoint now includes database connectivity status
- **Input Validation**: Comprehensive validation for all API request parameters (time ranges, pagination, filters)
- **End-to-End Tests**: Full integration tests for API → Service → Repository → Database flow
- **Log Security**: Sensitive credentials (accountID, patientID) are now masked in logs

### Changed
- **BREAKING**: Data endpoints require `/v1` prefix:
  - `GET /measurements/latest` → `GET /v1/measurements/latest`
  - `GET /measurements` → `GET /v1/measurements`
  - `GET /measurements/stats` → `GET /v1/measurements/stats`
  - `GET /sensors` → `GET /v1/sensors`
- Monitoring endpoints (`/health`, `/metrics`) remain at root level (unversioned)
- Configuration loading is now centralized with proper validation and error handling
- Magic numbers replaced with descriptive domain constants

### Fixed
- Mock repository implementation now includes all required methods
- Tests now compile and pass successfully
- Improved error messages for invalid API parameters

### Security
- Credentials and tokens are masked in debug logs
- API input validation prevents invalid time ranges and parameters
- Database password is now required when using PostgreSQL

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
