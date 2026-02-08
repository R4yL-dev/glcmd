# glcmd

## Overview

**Version**: 0.7.1
**Date**: 2026-02-08

glcmd is a glucose monitoring toolkit for retrieving and monitoring blood glucose data from the LibreView API using LibreLinkUp follower account credentials. It consists of two binaries:

- **glcore** — Daemon that polls the LibreView API, stores measurements in a local database, and exposes an HTTP REST API
- **glcli** — Command-line client for querying glucose data, sensor information, and statistics from the glcore API

## Features

- Continuous glucose monitoring via LibreView API follower account
- Daemon mode with automatic polling (default: 5 minutes)
- CLI client with Cobra-based subcommands, shell completion, and JSON output
- Local SQLite database persistence
- Automatic sensor change detection with unresponsive sensor detection
- HTTP REST API with 9 endpoints (health, metrics, measurements, statistics, sensors, sensor history, sensor stats, SSE stream)
- Real-time streaming via Server-Sent Events (SSE)
- Sensor lifecycle tracking with expiration, duration, and status fields
- All-time statistics support with SQL-based calculations
- Robust architecture with retry logic, transactions, and graceful shutdown
- Structured logging with configurable format (text/JSON) and level
- Open-source under MIT license

## Prerequisites

- Go 1.24.1 or higher
- CGO enabled (required for SQLite driver)
- GCC or compatible C compiler
  - Linux: `build-essential` package
  - macOS: Xcode Command Line Tools

Tested on Linux. Should work on macOS with prerequisites installed.

## Configuration

### Required Credentials

Set the following environment variables with LibreLinkUp follower account credentials:

```bash
export GLCMD_EMAIL='follower@example.com'
export GLCMD_PASSWORD='your_password'
```

**Important**: Credentials must be from a LibreLinkUp follower account, not the primary patient account from the Libre 3 app.

### Optional Configuration

Daemon settings:
- `GLCMD_FETCH_INTERVAL`: Polling interval (default: `5m`)
- `GLCMD_API_PORT`: HTTP API port (default: `8080`)

Logging settings:
- `GLCMD_LOG_FORMAT`: Log output format (`text` or `json`, default: `text`)
- `GLCMD_LOG_LEVEL`: Log verbosity (`debug`, `info`, `warn`, `error`, default: `info`)

Database settings:
- `GLCMD_DB_PATH`: SQLite database path (default: `./data/glcmd.db`)
- `GLCMD_DB_LOG_LEVEL`: GORM log level (default: `warn`)

For complete configuration options, see [Environment Variables documentation](docs/ENV_VARS.md).

## Installation

### Quick Start

```bash
# Set credentials
export GLCMD_EMAIL='follower@example.com'
export GLCMD_PASSWORD='password'

# Clone and build
git clone https://github.com/R4yL-dev/glcmd.git
cd glcmd
make

# Create data directory
mkdir -p data

# Run daemon
./bin/glcore

# Query data with CLI (in another terminal)
./bin/glcli
./bin/glcli stats --period 7d
./bin/glcli sensor
```

### Build Commands

```bash
# Build both binaries (glcore + glcli) to bin/
make

# Build individually
make build-glcore
make build-glcli

# Run directly
make run-glcore
make run-glcli

# Install to /usr/local/bin
make install

# Run tests
make test

# Clean build artifacts
make clean
```

## Usage

### Daemon (glcore)

glcore runs the monitoring daemon. Start it with:

```bash
./bin/glcore
```

The daemon:
- Polls LibreView API every 5 minutes (configurable via `GLCMD_FETCH_INTERVAL`)
- Stores measurements in SQLite database
- Detects sensor changes automatically and tracks unresponsive sensors
- Imports historical data on first run
- Persists data across restarts
- Exposes HTTP API on port 8080
- Logs to stderr with configurable format and level

### CLI Client (glcli)

glcli queries data from a running glcore instance:

```bash
# Current glucose reading
./bin/glcli

# Glucose statistics (7 days, 30 days, all-time)
./bin/glcli stats --period 7d
./bin/glcli stats --period all

# Glucose history
./bin/glcli history --period 24h
./bin/glcli history --start 2026-01-01 --end 2026-01-31

# Current sensor info
./bin/glcli sensor

# Sensor history and stats
./bin/glcli sensor history
./bin/glcli sensor stats

# GMI (Glucose Management Indicator)
./bin/glcli gmi

# Stream real-time events
./bin/glcli watch
./bin/glcli watch --only glucose
./bin/glcli watch --json

# JSON output for scripting
./bin/glcli --json
./bin/glcli --json stats --period 7d

# Custom API URL
./bin/glcli --api-url http://remote:8080 stats
# Or via environment variable
export GLCMD_API_URL=http://remote:8080
```

Shell completion is available via `glcli completion bash/zsh/fish/powershell`.

## HTTP API

glcore exposes a REST API on port 8080 for programmatic access to glucose data.

All data endpoints are versioned with the `/v1` prefix for API stability.

### Available Endpoints

**Monitoring endpoints** (unversioned):
- `GET /health` - Daemon and database health status with data freshness
- `GET /metrics` - Runtime metrics (uptime, memory, goroutines, SSE, DB pool)

**Data endpoints** (versioned):
- `GET /v1/glucose/latest` - Most recent glucose reading
- `GET /v1/glucose` - Paginated glucose measurements with filters
- `GET /v1/glucose/stats` - Glucose statistics with time-in-range analysis
- `GET /v1/sensor/latest` - Current active sensor information
- `GET /v1/sensor` - Paginated sensor list with date filters
- `GET /v1/sensor/stats` - Sensor lifecycle statistics with date filters

### Example API Calls

```bash
# Health check
$ curl http://localhost:8080/health | jq
{
  "data": {
    "status": "healthy",
    "timestamp": "2026-01-03T02:33:12Z",
    "uptime": "1h16m25s",
    "consecutiveErrors": 0,
    "lastFetchTime": "2026-01-03T02:31:47Z",
    "databaseConnected": true,
    "dataFresh": true,
    "fetchInterval": "5m0s"
  }
}

# Runtime metrics
$ curl http://localhost:8080/metrics | jq
{
  "data": {
    "uptime": "1h16m36s",
    "goroutines": 9,
    "memory": {
      "allocMB": 2,
      "totalAllocMB": 25,
      "sysMB": 19,
      "numGC": 42
    },
    "runtime": {
      "version": "go1.25.5",
      "os": "linux",
      "arch": "amd64"
    },
    "sse": {
      "enabled": true,
      "subscribers": 0
    },
    "database": {
      "openConnections": 1,
      "inUse": 0,
      "idle": 1,
      "waitCount": 0,
      "waitDuration": "0s"
    }
  }
}

# Latest glucose measurement
$ curl http://localhost:8080/v1/glucose/latest | jq
{
  "data": {
    "timestamp": "2026-01-03T02:31:27Z",
    "value": 5.6,
    "valueInMgPerDl": 100,
    "trendArrow": 3,
    "measurementColor": 1,
    "isHigh": false,
    "isLow": false
  }
}

# Sensor information
$ curl http://localhost:8080/v1/sensor/latest | jq
{
  "data": {
    "serialNumber": "ABC123XYZ",
    "activation": "2025-12-28T18:02:35Z",
    "expiresAt": "2026-02-10T18:02:35Z",
    "sensorType": 4,
    "durationDays": 14,
    "daysRemaining": 10.5,
    "daysElapsed": 3.5,
    "isActive": true,
    "status": "active",
    "isExpired": false,
    "isUnresponsive": false
  }
}

# All-time statistics
$ curl http://localhost:8080/v1/glucose/stats | jq

# Statistics for last 7 days
START=$(date -u -d '7 days ago' +%Y-%m-%dT%H:%M:%SZ)
END=$(date -u +%Y-%m-%dT%H:%M:%SZ)
curl "http://localhost:8080/v1/glucose/stats?start=$START&end=$END" | jq
```

For complete API documentation with all parameters and response formats, see [API Reference](docs/API.md).

## Documentation

- [API Reference](docs/API.md) - Complete HTTP API documentation
- [Architecture Documentation](docs/ARCHITECTURE.md) - Architectural overview and design patterns
- [Environment Variables](docs/ENV_VARS.md) - Comprehensive configuration reference
- [Documentation Index](docs/README.md) - Quick start guide and documentation overview
- [CHANGELOG](CHANGELOG.md) - Version history and release notes

## License

This project is licensed under the [MIT](LICENSE) license.

## Disclaimer

This tool is provided for informational and personal use only. It is not a certified medical device and should not be used to make health-related decisions. Use at your own risk.
