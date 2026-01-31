# glcmd Documentation

**Version**: 0.4.0
**Updated**: 2026-01-31

## Overview

This directory contains comprehensive documentation for the glcmd project, a LibreView glucose monitoring toolkit consisting of a daemon (`glcore`) and a CLI client (`glcli`).

## Documentation Files

### [API.md](API.md)

HTTP API reference documentation:
- Complete endpoint specifications
- Request/response formats
- Query parameters and validation
- Error handling
- Pagination and filtering
- Example curl commands
- Complete usage script

**Read this if you want to**:
- Integrate with the glcmd API programmatically
- Query glucose data via HTTP
- Monitor daemon health and metrics
- Build custom dashboards or integrations

### [ARCHITECTURE.md](ARCHITECTURE.md)

Complete architectural overview covering:
- Layered architecture design (domain, persistence, repository, service, daemon)
- Design patterns (Repository, Service, Unit of Work, Retry)
- Database schema and migrations
- Testing strategy and coverage
- Performance considerations
- Future migration path to PostgreSQL

**Read this if you want to**:
- Understand the codebase structure
- Learn about the design decisions
- Contribute to the project
- Extend functionality with new features

### [ENV_VARS.md](ENV_VARS.md)

Environment variable configuration reference:
- Database configuration (SQLite and PostgreSQL)
- Connection pooling settings
- Retry configuration
- Application settings
- Configuration examples for different environments
- Security best practices
- Troubleshooting guide

**Read this if you want to**:
- Configure glcmd for your environment
- Deploy glcmd in production
- Migrate from SQLite to PostgreSQL
- Troubleshoot connection or performance issues
- Set up containerized deployment

## Quick Start

### Development Setup

1. **Clone and build**:
   ```bash
   git clone <repository>
   cd glcmd
   make
   ```

2. **Create data directory**:
   ```bash
   mkdir -p data
   ```

3. **Set environment variables**:
   ```bash
   export GLCMD_EMAIL=your-email@example.com
   export GLCMD_PASSWORD=your-password
   export GLCMD_DB_TYPE=sqlite
   export GLCMD_DB_PATH=./data/glcmd.db
   export GLCMD_DB_LOG_LEVEL=info
   export GLCMD_API_PORT=8080
   ```

4. **Run daemon**:
   ```bash
   ./bin/glcore
   ```

5. **Query data with CLI** (in another terminal):
   ```bash
   ./bin/glcli
   ./bin/glcli stats --period 7d
   ./bin/glcli sensor
   ```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests with race detector
make test-race
```

### Production Deployment

See [ENV_VARS.md](ENV_VARS.md) for production configuration examples including:
- Systemd service configuration
- Docker Compose setup
- PostgreSQL migration guide
- Security best practices

## Architecture at a Glance

```
cmd/glcore (daemon entry point)
    ↓
internal/daemon (API polling)
    ↓
internal/api (unified HTTP server on port 8080)
    ↓
internal/service (business logic)
    ↓
internal/repository (data access)
    ↓
internal/persistence (database)
    ↓
internal/domain (models)

cmd/glcli (CLI entry point, Cobra)
    ↓
internal/cli (HTTP client + formatters)
```

**Key Features**:
- Two binaries: `glcore` (daemon) and `glcli` (CLI client)
- Unified HTTP API server on port 8080 with 8 endpoints
- CLI with Cobra subcommands, shell completion, and JSON output
- SQLite with WAL mode (current), PostgreSQL ready (future)
- ACID transactions via Unit of Work
- Automatic retry with exponential backoff
- Context-based timeout enforcement
- Comprehensive test coverage

## Key Concepts

### Repository Pattern
Data access abstraction that hides GORM implementation details. All repositories support transaction context propagation.

### Service Layer
Business logic and transaction orchestration. Services use Unit of Work for multi-step atomic operations.

### Unit of Work
Transaction management pattern using context propagation. Ensures ACID properties for complex operations.

### Retry Logic
Exponential backoff for transient database errors (locks, timeouts). Configurable via environment variables.

## Database

### Current: SQLite
- Single-file database at `./data/glcmd.db`
- WAL mode for better concurrency
- Auto-migrations via GORM
- Suitable for single-instance deployments

### Future: PostgreSQL
Architecture supports easy migration:
- Update environment variables
- No code changes required
- GORM abstracts database differences
- Containerization ready

## Contributing

When contributing to glcmd:

1. **Understand the architecture**: Read [ARCHITECTURE.md](ARCHITECTURE.md) first
2. **Follow patterns**: Use existing patterns (Repository, Service, UoW)
3. **Write tests**: Focus on critical paths and business logic
4. **Update documentation**: Keep docs in sync with code changes

### Adding New Features

**New Domain Model**:
1. Add to `internal/domain/` with GORM tags
2. Create repository interface in `internal/repository/interfaces.go`
3. Implement repository in `internal/repository/`
4. Add to auto-migration in `cmd/glcore/main.go`
5. Write tests for critical paths

**New Service**:
1. Define interface in `internal/service/interfaces.go`
2. Implement service with constructor injection
3. Wire up in `cmd/glcore/main.go`
4. Write tests with mocks for repositories

**New Repository Method**:
1. Add to interface in `internal/repository/interfaces.go`
2. Implement in concrete repository
3. Support transaction context via `txOrDefault()`
4. Write integration test with in-memory SQLite

## Troubleshooting

### Common Issues

**Database locked errors**:
- Ensure `DB_MAX_OPEN_CONNS=1` for SQLite
- Check retry configuration
- See [ENV_VARS.md](ENV_VARS.md) troubleshooting section

**Build failures**:
- SQLite driver requires CGO: `CGO_ENABLED=1`
- Install GCC on Linux: `apt-get install build-essential`
- Install Xcode Command Line Tools on macOS

**Connection errors**:
- Verify database path directory exists
- Check file permissions
- For PostgreSQL: verify credentials and network connectivity

## Additional Resources

- **GORM Documentation**: https://gorm.io/docs/
- **SQLite WAL Mode**: https://www.sqlite.org/wal.html
- **Go Context**: https://go.dev/blog/context
- **Testing in Go**: https://go.dev/doc/tutorial/add-a-test

## License

See main repository LICENSE file.
