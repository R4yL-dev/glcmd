# Changelog

All notable changes to glcmd are documented here.

## [0.7.1] - 2026-02-08

### Added
- **Health**: `dataFresh` field indicating whether glucose data is current (stale after 2x fetch interval)
- **Health**: `fetchInterval` field showing the configured polling interval
- **Metrics**: `sse` section with `enabled` status and active `subscribers` count
- **Metrics**: `database` section with connection pool statistics (open, in-use, idle, wait count/duration)

### Breaking Changes
- **Sensor status**: `"ended"` and `"expired"` statuses replaced by `"stopped"`
- **Sensor response**: Removed fields `isActive`, `isExpired`, `isUnresponsive`, `daysPastExpiry`
- **Sensor stats**: `endedSensors` renamed to `completedSensors`
- **ElapsedDays**: Now bounded by `ExpiresAt` for expired sensors (no longer grows past sensor lifetime)

### Changed
- Health status degrades to `degraded` when data becomes stale (no successful fetch within 2x interval)
- Sensor dates in CLI now display in local timezone instead of UTC

## [0.7.0] - 2026-02-08

### Added
- **Real-time Streaming**: Server-Sent Events (SSE) endpoint at `GET /v1/stream`
- **Event Broker**: Pub/sub event system for glucose and sensor changes
- **CLI**: `glcli watch` command for streaming real-time events
- **SSE Type Filtering**: `?types=glucose,sensor` query parameter for selective streaming
- **Keepalive**: Automatic 30-second heartbeat for SSE connection health

## [0.6.0] - 2026-02-08

### Breaking Changes
- **API**: `/v1/measurements` renamed to `/v1/glucose`
- **API**: `/v1/measurements/latest` renamed to `/v1/glucose/latest`
- **API**: `/v1/measurements/stats` renamed to `/v1/glucose/stats`
- **API**: `/v1/sensors` renamed to `/v1/sensor` (singular, now returns paginated list)
- **API**: `/v1/sensors/history` removed (merged into `/v1/sensor`)
- **API**: `/v1/sensors/stats` renamed to `/v1/sensor/stats`
- **CLI**: `--last` flag renamed to `--period` for `history` commands

### Added
- **API**: `GET /v1/sensor/latest` endpoint for current active sensor
- **API**: `start`/`end` query parameters for `/v1/sensor/stats`
- **CLI**: Dynamic period values (e.g., `--period 3d`, `--period 6w`, `--period 2m`)
- **CLI**: `--period`, `--start`, `--end` flags for `sensor stats` command

### Changed
- Internal: `MeasurementRepository` renamed to `GlucoseRepository`
- Internal: Centralized period parsing in `internal/utils/periodparser`
- Internal: Unified API structure for glucose and sensor endpoints

## [0.5.0] - 2026-02-07

### Added
- **Configurable Logging**: New environment variables for log configuration
  - `GLCMD_LOG_FORMAT`: Output format (`text` or `json`, default: `text`)
  - `GLCMD_LOG_LEVEL`: Log level (`debug`, `info`, `warn`, `error`, default: `info`)
- **Debug Logging**: Verbose logging for LibreView API client operations
- **Performance**: Glucose targets are now cached to avoid redundant database saves

### Changed
- Daemon logs now output to stderr instead of files
- Harmonized API request logging across components
- `SaveMeasurement` returns an `inserted` flag for tracking new vs. duplicate measurements

### Removed
- **BREAKING**: `GLCMD_DISPLAY_INTERVAL` environment variable removed
- **BREAKING**: `GLCMD_ENABLE_EMOJIS` environment variable removed
- Periodic display of glucose readings in daemon (use CLI or API instead)

## [0.4.0] - 2026-01-31

### Added
- **CLI Tool (`glcli`)**: New command-line client built with Cobra for querying the glcore API
  - `glcli` / `glcli glucose` — Display current glucose reading
  - `glcli glucose history` / `glcli history` — Browse historical measurements with filters
  - `glcli glucose stats` / `glcli stats` — View glucose statistics (today, 7d, 14d, 30d, 90d, all)
  - `glcli sensor` — Show current sensor information with status and lifecycle details
  - `glcli sensor history` — Browse past sensors with pagination and date filters
  - `glcli sensor stats` — View sensor lifecycle statistics (average duration, min/max, etc.)
  - Global `--json` flag for machine-readable output
  - Global `--api-url` flag and `GLCMD_API_URL` environment variable
  - Shell completion support (`glcli completion bash/zsh/fish/powershell`)
- **Sensor History Endpoint**: `GET /v1/sensors/history` — Paginated sensor list with start/end date filters
- **Sensor Statistics Endpoint**: `GET /v1/sensors/stats` — Sensor lifecycle statistics (total, completed, average duration, min/max)
- **Unresponsive Sensor Detection**: Sensors with no recent measurements are flagged as unresponsive
- **Last Measurement Tracking**: `lastMeasurementAt` field on sensor responses tracks most recent reading
- **All-Time Statistics**: `GET /v1/measurements/stats` now supports all-time queries when `start`/`end` are omitted

### Changed
- **BREAKING**: Daemon binary renamed from `glcmd` to `glcore` (`cmd/glcmd` → `cmd/glcore`)
- **BREAKING**: Sensor API response restructured with domain fields:
  - Removed raw fields (`sn`, `a`, `w`, `pt`, `s`, `lj`, `warranty_days`, `is_active`)
  - Added domain fields: `serialNumber`, `activation`, `expiresAt`, `endedAt`, `lastMeasurementAt`, `sensorType`, `durationDays`, `daysRemaining`, `daysElapsed`, `actualDays`, `daysPastExpiry`, `isActive`, `status`, `isExpired`, `isUnresponsive`
- **BREAKING**: Statistics endpoint `start`/`end` parameters are now optional (all-time if omitted); 90-day limit removed
- Sensor lifecycle model refactored: `IsActive` boolean replaced by `EndedAt` timestamp, added `ExpiresAt` and `DurationDays`
- Statistics calculations migrated from Go to SQL for better performance
- Makefile updated with `build-glcore`, `build-glcli`, `run-glcore`, `run-glcli` targets

### Fixed
- Removed duplicate log entry in daemon

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
