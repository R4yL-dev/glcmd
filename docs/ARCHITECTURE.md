# Architecture Documentation

**Version**: 0.7.0
**Updated**: 2026-02-08
**For**: glcmd glucose monitoring toolkit

## Overview

glcmd is a LibreView glucose monitoring toolkit built with a clean, layered architecture designed for maintainability and future scalability. It consists of two binaries: **glcore** (daemon) and **glcli** (CLI client). The daemon polls the LibreView API every 5 minutes and persists glucose measurements, sensor configurations, and user preferences to a SQLite database. The CLI client queries the daemon's HTTP API.

## Architecture Layers

```
┌─────────────────────────────────────┐
│           cmd/glcore                │  Daemon entry point, DI
└─────────────┬───────────────────────┘
              │
┌─────────────▼───────────────────────┐
│      internal/daemon                │  Orchestration, API polling
└─────────────┬───────────────────────┘
              │
┌─────────────▼───────────────────────┐
│      internal/service               │  Business logic, transactions
│  - GlucoseService                   │
│  - SensorService                    │
│  - ConfigService                    │
└─────────────┬───────────────────────┘
              │
┌─────────────▼───────────────────────┐
│     internal/repository             │  Data access abstraction
│  - MeasurementRepository            │
│  - SensorRepository                 │
│  - UserRepository                   │
│  - DeviceRepository                 │
│  - TargetsRepository                │
│  - UnitOfWork                       │
└─────────────┬───────────────────────┘
              │
┌─────────────▼───────────────────────┐
│    internal/persistence             │  Database infrastructure
│  - Database connection              │
│  - Retry logic                      │
│  - Configuration                    │
└─────────────┬───────────────────────┘
              │
┌─────────────▼───────────────────────┐
│       internal/domain               │  Domain models (GORM-tagged)
│  - GlucoseMeasurement               │
│  - SensorConfig                     │
│  - UserPreferences                  │
│  - DeviceInfo                       │
│  - GlucoseTargets                   │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│           cmd/glcli                 │  CLI entry point (Cobra)
└─────────────┬───────────────────────┘
              │
┌─────────────▼───────────────────────┐
│       internal/cli                  │  HTTP client + formatters
│  - Client (API consumer)            │
│  - Formatters (table/text output)   │
│  - Models (response types)          │
└─────────────────────────────────────┘
```

## Layer Responsibilities

### 1. Domain Layer (`internal/domain`)

Contains domain models with GORM tags for ORM mapping. These models are shared across all layers.

**Design Decision**: Unified models with GORM tags (no separate persistence models) for simplicity. Models include both database fields (GORM tags) and JSON serialization tags.

**Key Models**:
- `GlucoseMeasurement`: Glucose readings with unique timestamp constraint
- `SensorConfig`: CGM sensor metadata with active status tracking
- `UserPreferences`: User settings (units, targets, alarms)
- `DeviceInfo`: Device information
- `GlucoseTargets`: Glucose target ranges

### 2. Persistence Layer (`internal/persistence`)

Manages database connections, configuration, and infrastructure concerns.

**Components**:
- `Database`: GORM connection wrapper with health checks
- `DatabaseConfig`: Environment-driven configuration
- `RetryConfig`: Exponential backoff retry logic for database locks
- `ExecuteWithRetry()`: Retry wrapper for transient database errors

**Database Configuration**:
- SQLite with WAL (Write-Ahead Logging) mode for better concurrency
- Busy timeout: 5000ms
- Connection pooling: MaxOpenConns=1 (SQLite single writer limitation)
- Auto-migrations on startup via GORM

### 3. Repository Layer (`internal/repository`)

Implements data access patterns using GORM. All repositories support transaction context propagation.

**Pattern**: Context-based transactions using `context.Value()` to pass `*gorm.DB` transaction instances. Helper function `txOrDefault()` extracts transaction from context or falls back to default DB instance.

**Repositories**:
- `MeasurementRepository`: CRUD for glucose measurements with duplicate prevention
- `SensorRepository`: Sensor configuration with upsert behavior
- `UserRepository`: User preferences management
- `DeviceRepository`: Device information storage
- `TargetsRepository`: Glucose targets configuration
- `UnitOfWork`: Transaction management interface

**Key Features**:
- ON CONFLICT DO NOTHING for duplicate measurements (unique timestamp constraint)
- ON CONFLICT DO UPDATE for sensor configuration (upsert on serial number)
- Transaction context propagation via `txOrDefault(ctx, db)`
- Error wrapping for better debugging

### 4. Service Layer (`internal/service`)

Business logic and transaction orchestration.

**Services**:

#### GlucoseService
- Saves glucose measurements with retry logic
- Retrieves latest measurement
- Queries measurements by time range
- Performance logging for all operations

#### SensorService
- **Critical Business Logic**: `HandleSensorChange()` detects sensor changes atomically
  - Checks for existing active sensor
  - Deactivates old sensor if serial number changed
  - Saves new sensor configuration
  - All operations in single transaction (ACID guarantee)

#### ConfigService
- Manages user preferences
- Handles device information
- Manages glucose targets

**Pattern**: All services receive repositories and UnitOfWork via constructor injection. Services use UnitOfWork for multi-step operations requiring atomicity.

### 5. Daemon Layer (`internal/daemon`)

Orchestrates API polling and data persistence.

**Responsibilities**:
- Polls LibreView API every 5 minutes (configurable)
- Authenticates with LibreView (handles token expiration)
- Transforms API responses to domain models
- Delegates persistence to services

**Context Management**:
- All service calls include context.WithTimeout (5 seconds)
- Graceful shutdown via context cancellation

### 6. API Layer (`internal/api`)

Unified HTTP API server providing programmatic access to glucose data.

**Responsibilities**:
- Serves unified REST API on port 8080 (configurable via `GLCMD_API_PORT`)
- Exposes 9 endpoints: health, metrics, latest measurement, measurements list, statistics, sensors, sensor history, sensor stats, SSE stream
- Applies middleware: logging, recovery, timeout enforcement (REST endpoints only)
- Delegates data access to services
- Formats responses as consistent JSON with domain-level field names
- Provides real-time event streaming via SSE

**Integration**:
- Started alongside daemon in `cmd/glcore/main.go`
- Shares service layer with daemon
- Independent lifecycle management
- See [API.md](API.md) for complete endpoint specification

### 7. CLI Layer (`cmd/glcli` + `internal/cli`)

Command-line client for querying the glcore API.

**Architecture**:
- `cmd/glcli/cmd/` — Cobra command definitions (root, glucose, sensor, stats, history, version, completion)
- `internal/cli/client.go` — HTTP client that consumes the glcore REST API
- `internal/cli/formatter.go` — Text formatters for terminal display
- `internal/cli/models.go` — Response type definitions for JSON deserialization

**Features**:
- Cobra-based subcommand tree with shell completion
- Global `--json` flag for machine-readable output
- Global `--api-url` flag (default from `GLCMD_API_URL` or `http://localhost:8080`)
- Formatted table/text output for glucose readings, statistics, and sensor info

**Commands**:
- `glcli` / `glcli glucose` — Current glucose reading
- `glcli history` / `glcli glucose history` — Historical measurements
- `glcli stats` / `glcli glucose stats` — Glucose statistics
- `glcli gmi` / `glcli glucose gmi` — Glucose Management Indicator (estimated A1C)
- `glcli sensor` — Current sensor info
- `glcli sensor history` — Past sensors
- `glcli sensor stats` — Sensor lifecycle statistics
- `glcli watch` — Real-time event streaming
- `glcli version` — Version information
- `glcli completion` — Shell completion scripts

### 8. Event System (`internal/events`)

Real-time event distribution using a pub/sub pattern.

**Components**:
- `Broker` — Central event hub managing subscriptions and distribution
- `Event` — Generic event wrapper with type and data payload
- `Subscriber` — Client subscription with optional type filtering

**Event Flow**:
1. Services publish events after successful operations (new measurement, sensor change)
2. Broker distributes events to all matching subscribers (non-blocking)
3. SSE handler forwards events to connected HTTP clients
4. CLI/Frontend receives and displays events in real-time

**Heartbeat**:
- Broker sends `keepalive` events every 30 seconds
- Allows detection of dead connections
- Clients can use heartbeats to verify connection health

**Thread Safety**:
- All broker operations are thread-safe using RWMutex
- Non-blocking publish prevents slow subscribers from affecting others
- Channel buffer size configurable (default: 10 events)

## Recent Changes (v0.7.0)

### Real-time Streaming (SSE)
- New `internal/events` package with pub/sub event broker
- SSE endpoint at `/v1/stream` for real-time event streaming
- New `glcli watch` command for CLI streaming
- Services publish events on new measurements and sensor changes
- Heartbeat mechanism (30s interval) for connection health monitoring
- Type filtering support (glucose, sensor, keepalive)

## Previous Changes (v0.5.0)

### Daemon Simplification
- Removed periodic display of glucose readings (display ticker)
- Removed emoji configuration support
- All glucose data accessed via REST API or CLI client

### Logging Improvements
- Console logger with configurable format (`text` or `json`) via `GLCMD_LOG_FORMAT`
- Configurable log level (`debug`, `info`, `warn`, `error`) via `GLCMD_LOG_LEVEL`
- Logs output to stderr for container compatibility
- Debug logging for LibreView API client operations
- Harmonized API request logging across components

### Performance Optimizations
- Glucose targets cached to avoid redundant database saves
- `SaveMeasurement` returns `inserted` flag for tracking new vs. duplicate measurements

## Changes in v0.4.0

### CLI Client
- New `glcli` binary built with Cobra for querying the glcore API
- `internal/cli` package with HTTP client, formatters, and response models
- Subcommands for glucose, sensor, stats, and history with filters and pagination
- Global `--json` flag and `--api-url` / `GLCMD_API_URL` configuration
- Shell completion support (bash, zsh, fish, powershell)

### Sensor Lifecycle Refactoring
- Replaced `IsActive` boolean with `EndedAt` timestamp for accurate lifecycle tracking
- Added `ExpiresAt` and `DurationDays` fields computed from sensor metadata
- Added `LastMeasurementAt` field for tracking most recent measurement per sensor
- Unresponsive sensor detection when measurements stop arriving
- New API response format with domain-level fields (`SensorResponse` struct)

### New API Endpoints
- `GET /v1/sensors/history` — Paginated sensor list with start/end date filters
- `GET /v1/sensors/stats` — Sensor lifecycle statistics (total, completed, durations)

### Statistics Improvements
- Calculations migrated from Go to SQL for better performance
- `start`/`end` parameters now optional — returns all-time stats if omitted
- Removed 90-day limit on time ranges

### Binary Rename
- Daemon binary renamed from `glcmd` to `glcore` (`cmd/glcmd` → `cmd/glcore`)
- Makefile updated with `build-glcore`, `build-glcli`, `run-glcore`, `run-glcli` targets

### Previous Changes (v0.3.0)
- All data endpoints versioned with `/v1` prefix
- CORS middleware for web frontend compatibility
- Centralized configuration in `internal/config` package
- Domain constants for measurement colors, trend arrows, and glucose units
- End-to-end tests covering full API → Service → Repository → Database flow
- Credentials and tokens masked in logs

### Previous Changes (v0.2.0)
- Consolidated health/metrics servers into unified server on port 8080
- Comprehensive test suite for transaction safety and data integrity
- Structured logging with slog, circuit breaker pattern, database health checks

## Key Patterns

### Unit of Work Pattern

Transaction management via context propagation. The `UnitOfWork` interface provides `ExecuteInTransaction()` which:
1. Begins a GORM transaction
2. Stores transaction in context via `context.WithValue()`
3. Executes business logic function
4. Commits on success or rolls back on error

**Example**:
```go
err := uow.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
    // All repository operations within txCtx use the same transaction
    err := repo1.Save(txCtx, data1)
    if err != nil { return err }

    err = repo2.Save(txCtx, data2)
    if err != nil { return err }

    return nil
})
// Transaction automatically committed or rolled back
```

### Repository Pattern

Data access abstraction with interface-based contracts. Repositories hide GORM implementation details and provide clean domain-focused API.

**Transaction Support**: Repositories check context for transaction via `txOrDefault()`:
```go
func (r *Repository) Save(ctx context.Context, entity *domain.Entity) error {
    db := txOrDefault(ctx, r.db) // Use transaction if present, else default DB
    return db.Create(entity).Error
}
```

### Retry Pattern

Exponential backoff for transient database errors (locks, timeouts).

**Configuration**:
- MaxRetries: 3
- InitialBackoff: 100ms
- MaxBackoff: 500ms
- Multiplier: 2.0
- Progression: 100ms → 200ms → 400ms

**Retryable Errors**:
- "database is locked"
- "SQLITE_BUSY"
- "connection refused"
- "timeout"

**Non-retryable Errors**:
- `persistence.ErrNotFound`
- Validation errors
- Constraint violations

### Context Propagation

All layers respect context for:
- Timeout enforcement (5 seconds for service operations)
- Transaction propagation (via context.Value)
- Graceful shutdown (daemon context cancellation)

## Database Schema

### glucose_measurements
- `id`: Primary key
- `created_at`: Record creation timestamp
- `timestamp`: Measurement timestamp (unique index)
- `value`: Glucose value in mmol/L
- `value_in_mg_per_dl`: Glucose value in mg/dL
- `trend_arrow`: Trend direction (-2 to +2)
- Indexes: `idx_unique_timestamp` (unique), `idx_timestamp`

### sensor_configs
- `id`: Primary key
- `created_at`: Record creation timestamp
- `updated_at`: Record update timestamp
- `serial_number`: Sensor serial (unique index)
- `activation`: Sensor activation timestamp (index)
- `expires_at`: Expected expiration timestamp
- `ended_at`: Actual end timestamp (null if active)
- `last_measurement_at`: Most recent measurement timestamp
- `sensor_type`: Sensor type code
- `duration_days`: Expected sensor duration in days
- `detected_at`: First detection timestamp
- Indexes: `idx_serial` (unique), `idx_activation`

### user_preferences
- `id`: Primary key
- `user_id`: User identifier (unique)
- `units`: Glucose units (0=mmol/L, 1=mg/dL)
- `low_alarm_value`: Low alarm threshold
- `high_alarm_value`: High alarm threshold
- `alarms`: Enabled alarms (JSON array)

### device_infos
- `id`: Primary key
- `device_id`: Device identifier (unique)
- `family`: Device family
- `model`: Device model
- `country`: Country code

### glucose_targets
- `id`: Primary key
- `target_low`: Target low value
- `target_high`: Target high value

## Testing Strategy

Tests focus on critical paths and business logic correctness.

### Coverage Areas

1. **Transaction Management** (`uow_test.go`)
   - Commit on success
   - Rollback on error
   - Context propagation
   - Atomic multi-operation rollback

2. **Data Integrity** (`measurement_repo_test.go`)
   - Duplicate timestamp prevention
   - Time range queries
   - Ordering (newest first)

3. **Sensor Logic** (`sensor_repo_test.go`, `sensor_test.go`)
   - Upsert behavior
   - Active sensor detection
   - Sensor change detection with mocks
   - Transaction rollback scenarios

4. **Retry Logic** (`retry_test.go`)
   - Exponential backoff timing
   - Retryable vs non-retryable errors
   - Context cancellation
   - Max retries enforcement

**Test Database**: SQLite in-memory (`:memory:`) for fast, isolated integration tests.

**Philosophy**: Few useful tests over many trivial tests. Focus on critical business logic and data integrity.

## Database

### SQLite
- Single-file database: `./data/glcmd.db`
- WAL mode for concurrent reads during writes
- Auto-migrations via GORM
- Suitable for single-instance deployments

## Performance Considerations

### SQLite Optimizations
- WAL mode for concurrent reads during writes
- Busy timeout prevents immediate failures on locks
- Prepared statements via GORM (`PrepareStmt: true`)

### Connection Pooling
- MaxOpenConns: 1 (SQLite single writer)
- MaxIdleConns: 1
- ConnMaxLifetime: 1 hour

### Retry Logic
- Automatic retry for transient lock errors
- Exponential backoff prevents thundering herd
- Context-aware cancellation

### Logging
- Structured logging with slog to stderr
- Configurable output format (text or JSON) via `GLCMD_LOG_FORMAT`
- Configurable log level via `GLCMD_LOG_LEVEL`
- Duration tracking for all service operations
- Debug-level logging for API requests

## Error Handling

### Error Types
- `persistence.ErrNotFound`: Entity not found (not retryable)
- Database lock errors: Retryable with exponential backoff
- Constraint violations: Not retryable
- Context errors: Not retryable (timeout/cancellation)

### Error Wrapping
All errors wrapped with context using `fmt.Errorf()` and `%w` verb for error chain preservation.

### Error Logging
- Service layer logs all errors with structured context
- Duration tracking for performance analysis
- Transaction failures logged with rollback details

## Future Enhancements

### Under Consideration
- SQL migration tool (e.g., golang-migrate) for versioned schema changes
- Prometheus metrics export for monitoring integration
- Query result caching for frequently accessed data
- Batch insert optimization for historical data imports

### Backwards Compatibility
The layered architecture allows adding features without breaking changes:
- New repositories: Add interface + implementation
- New services: Constructor injection of new repositories
- New domain fields: GORM auto-migration handles schema updates
- Configuration changes: Environment variable based (no code changes)
