# glcmd

## Overview

**Version**: 0.3.0
**Date**: 2026-01-03

glcmd is a command-line daemon for retrieving and monitoring blood glucose data from the LibreView API using LibreLinkUp follower account credentials. It provides continuous glucose monitoring in the terminal with local SQLite persistence and an HTTP REST API for programmatic access.

The daemon polls the LibreView API at configurable intervals, stores measurements in a local database, and exposes glucose data through a unified HTTP API server.

## Features

- Continuous glucose monitoring via LibreView API follower account
- Daemon mode with automatic polling (default: 5 minutes)
- Local database persistence (SQLite default, PostgreSQL supported)
- Automatic sensor change detection with historical data import
- HTTP REST API with 6 endpoints (health, metrics, measurements, statistics, sensors)
- Robust architecture with retry logic, transactions, and graceful shutdown
- Structured logging with configurable output
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

Common daemon settings:
- `GLCMD_FETCH_INTERVAL`: Polling interval (default: `5m`)
- `GLCMD_DISPLAY_INTERVAL`: Display interval (default: `1m`)
- `GLCMD_ENABLE_EMOJIS`: Enable emoji display (default: `true`)
- `GLCMD_API_PORT`: HTTP API port (default: `8080`)

Database settings:
- `GLCMD_DB_TYPE`: Database type (`sqlite` or `postgres`, default: `sqlite`)
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
./bin/glcmd
```

### Build Commands

```bash
# Build binary to bin/glcmd
make

# Install to /usr/local/bin
make install

# Run tests
make test

# Clean build artifacts
make clean
```

## Usage

glcmd runs exclusively in daemon mode. Start the daemon with:

```bash
./bin/glcmd
```

### Example Output

```
time=2026-01-03T02:04:47.896+01:00 level=INFO msg="glcmd starting"
time=2026-01-03T02:04:47.920+01:00 level=INFO msg="database connected successfully" type=sqlite path=./data/glcmd.db
time=2026-01-03T02:04:47.925+01:00 level=INFO msg="services initialized successfully"
time=2026-01-03T02:04:47.930+01:00 level=INFO msg="daemon configuration loaded" fetchInterval=5m displayInterval=1m emojisEnabled=true
time=2026-01-03T02:04:47.935+01:00 level=INFO msg="unified API server started" port=8080
time=2026-01-03T02:04:47.940+01:00 level=INFO msg="daemon started, polling LibreView API every 5m"
time=2026-01-03T02:04:47.945+01:00 level=INFO msg="fetching new measurement"
time=2026-01-03T02:04:48.120+01:00 level=INFO msg="measurement fetched successfully"
time=2026-01-03T02:04:48.120+01:00 level=INFO msg="last measurement" value="5.7 mmol/L (102 mg/dL)" trend=➡️ status="🟢 Normal" timestamp="2026-01-03 02:04:28"
time=2026-01-03T02:05:47.837+01:00 level=INFO msg="last measurement" value="5.7 mmol/L (102 mg/dL)" trend=➡️ status="🟢 Normal" timestamp="2026-01-03 02:04:28"
time=2026-01-03T02:06:47.891+01:00 level=INFO msg="fetching new measurement"
time=2026-01-03T02:06:48.039+01:00 level=INFO msg="measurement fetched successfully"
time=2026-01-03T02:06:48.039+01:00 level=INFO msg="last measurement" value="5.6 mmol/L (100 mg/dL)" trend=➡️ status="🟢 Normal" timestamp="2026-01-03 02:06:27"
```

The daemon:
- Polls LibreView API every 5 minutes (configurable)
- Displays latest glucose reading every minute (configurable)
- Stores measurements in SQLite database
- Detects sensor changes automatically
- Imports historical data on first run
- Persists data across restarts
- Exposes HTTP API on port 8080

## HTTP API

The daemon exposes a REST API on port 8080 for programmatic access to glucose data.

All data endpoints are versioned with the `/v1` prefix for API stability.

### Available Endpoints

**Monitoring endpoints** (unversioned):
- `GET /health` - Daemon and database health status
- `GET /metrics` - Runtime metrics (uptime, memory, goroutines)

**Data endpoints** (versioned):
- `GET /v1/measurements/latest` - Most recent glucose reading
- `GET /v1/measurements` - Paginated measurements with filters
- `GET /v1/measurements/stats` - Statistics with time-in-range analysis
- `GET /v1/sensors` - Sensor information

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
    "lastFetchTime": "2026-01-03T02:31:47Z"
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
    }
  }
}

# Latest glucose measurement
$ curl http://localhost:8080/v1/measurements/latest | jq
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
$ curl http://localhost:8080/v1/sensors | jq
{
  "data": {
    "sensors": [
      {
        "serialNumber": "ABC123XYZ",
        "activation": "2025-12-28T18:02:35Z",
        "sensorType": 4,
        "warrantyDays": 60,
        "isActive": false
      }
    ]
  }
}

# Statistics for last 7 days
START=$(date -u -d '7 days ago' +%Y-%m-%dT%H:%M:%SZ)
END=$(date -u +%Y-%m-%dT%H:%M:%SZ)
curl "http://localhost:8080/measurements/stats?start=$START&end=$END" | jq
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
