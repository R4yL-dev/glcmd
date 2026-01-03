# Architecture Documentation

**Version**: 0.2.0
**Updated**: 2026-01-03
**For**: glcmd glucose monitoring daemon

## Overview

glcmd is a LibreView glucose monitoring daemon built with a clean, layered architecture designed for maintainability and future scalability. The application polls the LibreView API every 5 minutes and persists glucose measurements, sensor configurations, and user preferences to a SQLite database.

## Architecture Layers

```
┌─────────────────────────────────────┐
│           cmd/glcmd                 │  Entry point, dependency injection
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
- Periodic display of latest glucose reading

**Context Management**:
- All service calls include context.WithTimeout (5 seconds)
- Graceful shutdown via context cancellation

### 6. API Layer (`internal/api`)

Unified HTTP API server providing programmatic access to glucose data.

**Responsibilities**:
- Serves unified REST API on port 8080 (configurable via `GLCMD_API_PORT`)
- Exposes 6 endpoints: health, metrics, latest measurement, measurements list, statistics, sensors
- Applies middleware: logging, recovery, timeout enforcement
- Delegates data access to services
- Formats responses as consistent JSON

**Integration**:
- Started alongside daemon in main.go
- Shares service layer with daemon
- Independent lifecycle management
- See [API.md](API.md) for complete endpoint specification

## Recent Changes (v0.2.0)

### API Unification
- Consolidated separate health/metrics servers into single unified server
- All endpoints now on port 8080
- Simplifies deployment and monitoring configuration
- Reduces resource overhead

### Testing Improvements
- Comprehensive test suite for critical components
- Focus on transaction safety and data integrity
- Unit of Work transaction tests
- Repository integration tests with in-memory SQLite
- Retry logic verification

### Architecture Refinements
- Structured logging with slog throughout codebase
- Enhanced error handling with context propagation
- Database health checks on startup
- Circuit breaker pattern for API resilience

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
- `serial_number`: Sensor serial (unique index)
- `activation`: Sensor activation timestamp
- `device_id`: Device identifier
- `sensor_type`: Sensor type code
- `warranty_days`: Warranty duration
- `is_active`: Active status (index)
- `low_journey`: Low journey flag
- `detected_at`: First detection timestamp
- Indexes: `idx_serial` (unique), `idx_active`

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

## Migration Path

### Current: SQLite
- Single-file database: `./data/glcmd.db`
- Auto-migrations via GORM
- Suitable for single-instance deployments

### Future: PostgreSQL
The architecture is designed for easy migration:

1. **Configuration Change**: Set environment variables for PostgreSQL connection
2. **No Code Changes**: GORM abstracts database differences
3. **Migration Strategy**:
   - Export SQLite data
   - Import to PostgreSQL
   - Update configuration
   - Restart application

**Container Strategy**:
- App container: glcmd binary with PostgreSQL configuration
- DB container: PostgreSQL instance
- Docker Compose for orchestration

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
- Structured logging with slog
- Duration tracking for all service operations
- GORM query logging at WARN level (production)

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

### Planned for v0.3.0
- ASCII graph visualization of glucose trends in terminal
- Real-time notifications for critical glucose levels
- PostgreSQL migration for production deployments with migration guide

### Under Consideration
- SQL migration tool (e.g., golang-migrate) for versioned schema changes
- Prometheus metrics export for monitoring integration
- Read replicas support for PostgreSQL
- Connection pool tuning for high-traffic PostgreSQL deployments
- Query result caching for frequently accessed data
- Batch insert optimization for historical data imports

### Backwards Compatibility
The layered architecture allows adding features without breaking changes:
- New repositories: Add interface + implementation
- New services: Constructor injection of new repositories
- New domain fields: GORM auto-migration handles schema updates
- Configuration changes: Environment variable based (no code changes)
