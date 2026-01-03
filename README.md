# glcmd

## 🎯 About

**Version**: 0.1.1
**Date**: 2025-05-04

`glcmd` is a command-line tool designed to retrieve and display blood glucose information from the **LibreView API** using a **LibreLinkUp follower account**. It allows users to quickly and easily monitor their glucose levels directly in the terminal, without the need for proprietary apps.

This tool is ideal for people who want to have more control and flexibility over their glucose data, providing a simple, open-source alternative for tracking and displaying their measurements.

### 🌟 Key Features

- **Retrieve current glucose readings** from the LibreView API using a **follower account**.
- **Display glucose levels** in the terminal in a human-readable format (mmol/L).
- **Daemon mode** for continuous monitoring with automatic polling every 5 minutes.
- **Database persistence** with SQLite (PostgreSQL ready for containerized deployments).
- **Automatic sensor change detection** and historical data import.
- **HTTP REST API** with health checks, metrics, and comprehensive glucose data endpoints.
- **Robust architecture** with retry logic, transactions, and context-based timeout management.
- **Open-source**: freely available to use, modify, and contribute to.
- **Planned improvements**:
  - ASCII graph to visualize glucose trends in terminal.
  - Real-time notifications for critical glucose levels.

### 💡 Why `glcmd`?

Managing diabetes requires constant tracking and monitoring of glucose levels. `glcmd` was created to offer users a lightweight, no-frills tool to access their glucose data without being tied to a proprietary platform or app. It aims to give people more flexibility, transparency, and control over their health data in a simple command-line interface.

## 📦 Prerequisites

- **Go** 1.24.1 or higher
- **CGO** enabled (required for SQLite driver)
- **GCC** or compatible C compiler (Linux: `build-essential`, macOS: Xcode Command Line Tools)

> 🚨 This project has been tested on **Linux** and should work on **macOS** with the prerequisites installed.
> `make install` places the binary in `/usr/local/bin`. If this folder does not exist on macOS, simply compile it with `make` and move the binary to a folder included in your `PATH`.

## ⚙️ Setup

Before using `glcmd`, you need to configure your LibreView credentials.

### Required Credentials

These credentials must belong to a **follower account** — meaning an associated device account (not your primary patient account from the Libre 3 app).
The follower account must be added as an associated device in the Libre 3 application.
Using direct patient account credentials will not work.

Set the following environment variables:
- `GLCMD_EMAIL`: Your LibreView follower account email
- `GLCMD_PASSWORD`: Your LibreView follower account password

### Optional Configuration

For complete database and daemon configuration, see the [Environment Variables documentation](docs/ENV_VARS.md).

**Common daemon settings**:
- `GLCMD_FETCH_INTERVAL`: How often to fetch data from LibreView API (default: `5m`)
- `GLCMD_DISPLAY_INTERVAL`: How often to display latest measurement (default: `1m`)
- `GLCMD_ENABLE_EMOJIS`: Enable emoji display in logs (default: `true`)
- `GLCMD_API_PORT`: HTTP port for unified API server (default: `8080`)

**Database settings**:
- `GLCMD_DB_TYPE`: Database type (`sqlite` or `postgres`, default: `sqlite`)
- `GLCMD_DB_PATH`: Path to SQLite database (default: `./data/glcmd.db`)
- `GLCMD_DB_LOG_LEVEL`: GORM log level (`silent`, `error`, `warn`, `info`, default: `warn`)

## 🚀 Install & Usage

### Quick Start (One-time Query)

```bash
export GLCMD_EMAIL='<email>'
export GLCMD_PASSWORD='<password>'
git clone https://github.com/R4yL-dev/glcmd.git
cd glcmd
make
./bin/glcmd
🩸 7.7(mmol/L) 🡒
```

### Daemon Mode (Continuous Monitoring)

```bash
# Set credentials
export GLCMD_EMAIL='<email>'
export GLCMD_PASSWORD='<password>'

# Optional: Configure database
export GLCMD_DB_TYPE=sqlite
export GLCMD_DB_PATH=./data/glcmd.db
export GLCMD_DB_LOG_LEVEL=warn

# Create data directory
mkdir -p data

# Build and run daemon
make
./bin/glcmd daemon
```

**Daemon features**:
- Polls LibreView API every 5 minutes (configurable via `GLCMD_FETCH_INTERVAL`)
- Stores measurements, sensor info, and preferences in SQLite database
- Displays latest glucose reading every minute (configurable via `GLCMD_DISPLAY_INTERVAL`)
- Automatically detects sensor changes
- Imports historical data on first run
- Persists data across restarts
- **HTTP API server** on port 8080 with health, metrics, and data endpoints (see [API Documentation](#-api-documentation))
- Circuit breaker with automatic error recovery

### Build & Install

```bash
# Build only
make

# Build and install to /usr/local/bin
make install

# Run tests
make test

# Clean build artifacts
make clean
```

## 🌐 API Documentation

When running in daemon mode, `glcmd` provides a unified HTTP API server on port 8080 (configurable via `GLCMD_API_PORT`). All endpoints return JSON responses with consistent formatting and pass through logging, recovery, and timeout middlewares.

### Base URL

```
http://localhost:8080
```

### Response Format

All successful responses follow this structure:
```json
{
  "data": {
    // Response data here
  }
}
```

Error responses:
```json
{
  "error": {
    "code": 400,
    "message": "Error description"
  }
}
```

### Endpoints

#### 1. Health Check

**GET** `/health`

Returns the daemon health status.

**Response Codes:**
- `200 OK` - Service is healthy
- `503 Service Unavailable` - Service is degraded or unhealthy

**Response:**
```json
{
  "data": {
    "status": "healthy",
    "timestamp": "2025-01-03T10:30:00Z",
    "uptime": "2h15m30s",
    "consecutiveErrors": 0,
    "lastFetchError": "",
    "lastFetchTime": "2025-01-03T10:29:45Z"
  }
}
```

**Status Values:**
- `healthy` - All systems operational
- `degraded` - Some errors but still functional (1-2 consecutive errors)
- `unhealthy` - Service experiencing issues (3+ consecutive errors)

**Example:**
```bash
curl http://localhost:8080/health | jq
```

---

#### 2. Metrics

**GET** `/metrics`

Returns runtime metrics and system information.

**Response:**
```json
{
  "data": {
    "uptime": "2h15m30s",
    "goroutines": 12,
    "memory": {
      "allocMB": 8,
      "totalAllocMB": 145,
      "sysMB": 23,
      "numGC": 42
    },
    "runtime": {
      "version": "go1.21.5",
      "os": "linux",
      "arch": "amd64"
    },
    "process": {
      "pid": 1234
    }
  }
}
```

**Example:**
```bash
curl http://localhost:8080/metrics | jq
```

---

#### 3. Latest Measurement

**GET** `/measurements/latest`

Returns the most recent glucose measurement.

**Response:**
```json
{
  "data": {
    "id": 123,
    "timestamp": "2025-01-03T10:29:45Z",
    "factoryTimestamp": "2025-01-03T10:29:45Z",
    "value": 7.7,
    "valueInMgPerDl": 139,
    "trendArrow": 3,
    "trendMessage": "→",
    "measurementColor": 1,
    "glucoseUnits": 0,
    "isHigh": false,
    "isLow": false
  }
}
```

**Field Descriptions:**
- `value` - Glucose value in mmol/L
- `valueInMgPerDl` - Glucose value in mg/dL
- `trendArrow` - Trend indicator (1-5)
- `trendMessage` - Arrow symbol (↑, ↗, →, ↘, ↓)
- `measurementColor` - Color indicator (1=normal, 2=warning, 3=critical)
- `glucoseUnits` - Unit type (0=mmol/L, 1=mg/dL)

**Example:**
```bash
curl http://localhost:8080/measurements/latest | jq
```

---

#### 4. Measurements List

**GET** `/measurements`

Returns a paginated list of glucose measurements with optional filters.

**Query Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `limit` | integer | No | 100 | Number of results per page (1-1000) |
| `offset` | integer | No | 0 | Number of results to skip |
| `start` | string (RFC3339) | No | - | Filter measurements after this time |
| `end` | string (RFC3339) | No | - | Filter measurements before this time |
| `color` | integer | No | - | Filter by color (1=normal, 2=warning, 3=critical) |
| `type` | integer | No | - | Filter by type (0=historical, 1=current) |

**Response:**
```json
{
  "data": [
    {
      "id": 123,
      "timestamp": "2025-01-03T10:29:45Z",
      "value": 7.7,
      "valueInMgPerDl": 139,
      "trendArrow": 3,
      "measurementColor": 1
    }
  ],
  "pagination": {
    "limit": 100,
    "offset": 0,
    "total": 1542,
    "hasMore": true
  }
}
```

**Examples:**
```bash
# Get first 50 measurements
curl "http://localhost:8080/measurements?limit=50" | jq

# Get measurements from last 24 hours
START=$(date -u -d '24 hours ago' +%Y-%m-%dT%H:%M:%SZ)
END=$(date -u +%Y-%m-%dT%H:%M:%SZ)
curl "http://localhost:8080/measurements?start=$START&end=$END" | jq

# Get only warning/critical measurements
curl "http://localhost:8080/measurements?color=2" | jq
curl "http://localhost:8080/measurements?color=3" | jq

# Pagination example - get next page
curl "http://localhost:8080/measurements?limit=100&offset=100" | jq
```

---

#### 5. Statistics

**GET** `/measurements/stats`

Returns glucose statistics for a specified time period.

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `start` | string (RFC3339) | **Yes** | Start of time range |
| `end` | string (RFC3339) | **Yes** | End of time range (max 90 days from start) |

**Response:**
```json
{
  "data": {
    "period": {
      "start": "2025-01-01T00:00:00Z",
      "end": "2025-01-03T23:59:59Z"
    },
    "statistics": {
      "count": 864,
      "averageGlucose": 7.2,
      "minGlucose": 4.5,
      "maxGlucose": 12.3,
      "stdDev": 1.8,
      "cv": 25.0,
      "lowCount": 12,
      "normalCount": 800,
      "highCount": 52,
      "timeInRange": 92.59,
      "timeBelowRange": 1.39,
      "timeAboveRange": 6.02
    },
    "timeInRange": {
      "targetLowMgDl": 70,
      "targetHighMgDl": 180,
      "inRange": 92.59,
      "belowRange": 1.39,
      "aboveRange": 6.02
    },
    "distribution": {
      "low": 12,
      "normal": 800,
      "high": 52
    }
  }
}
```

**Field Descriptions:**
- `count` - Total number of measurements
- `averageGlucose` - Mean glucose value (mmol/L)
- `stdDev` - Standard deviation
- `cv` - Coefficient of variation (%)
- `timeInRange` - Percentage of time in target range
- `timeBelowRange` - Percentage of time below target
- `timeAboveRange` - Percentage of time above target

**Examples:**
```bash
# Get statistics for last 7 days
START=$(date -u -d '7 days ago' +%Y-%m-%dT%H:%M:%SZ)
END=$(date -u +%Y-%m-%dT%H:%M:%SZ)
curl "http://localhost:8080/measurements/stats?start=$START&end=$END" | jq

# Get statistics for a specific day
curl "http://localhost:8080/measurements/stats?start=2025-01-01T00:00:00Z&end=2025-01-01T23:59:59Z" | jq

# Get statistics for last 30 days
START=$(date -u -d '30 days ago' +%Y-%m-%dT%H:%M:%SZ)
END=$(date -u +%Y-%m-%dT%H:%M:%SZ)
curl "http://localhost:8080/measurements/stats?start=$START&end=$END" | jq
```

---

#### 6. Sensors

**GET** `/sensors`

Returns information about all sensors and the currently active sensor.

**Response:**
```json
{
  "data": {
    "active": {
      "sensorId": "ABC123XYZ",
      "sn": "0M0012345678",
      "a": 150,
      "w": 14400,
      "pt": 60,
      "s": false,
      "lj": false
    },
    "sensors": [
      {
        "sensorId": "ABC123XYZ",
        "sn": "0M0012345678",
        "a": 150,
        "w": 14400,
        "pt": 60,
        "s": false,
        "lj": false
      }
    ]
  }
}
```

**Field Descriptions:**
- `sensorId` - Unique sensor identifier
- `sn` - Sensor serial number
- `a` - Age in minutes
- `w` - Warmup time in seconds
- `pt` - Measurement interval in seconds
- `s` - Sensor started flag
- `lj` - Libreジャンプ flag

**Example:**
```bash
curl http://localhost:8080/sensors | jq
```

---

### Error Handling

All endpoints use consistent error handling:

**Validation Errors (400 Bad Request):**
```json
{
  "error": {
    "code": 400,
    "message": "limit must be at least 1"
  }
}
```

**Not Found (404):**
```json
{
  "error": {
    "code": 404,
    "message": "No measurements found"
  }
}
```

**Timeout (504 Gateway Timeout):**
```json
{
  "error": {
    "code": 504,
    "message": "Request timeout"
  }
}
```

**Internal Server Error (500):**
```json
{
  "error": {
    "code": 500,
    "message": "Internal server error"
  }
}
```

---

### Complete Example Script

```bash
#!/bin/bash

BASE_URL="http://localhost:8080"

# Check health
echo "=== Health Check ==="
curl -s "$BASE_URL/health" | jq

# Get metrics
echo -e "\n=== Metrics ==="
curl -s "$BASE_URL/metrics" | jq

# Get latest measurement
echo -e "\n=== Latest Measurement ==="
curl -s "$BASE_URL/measurements/latest" | jq

# Get last 10 measurements
echo -e "\n=== Last 10 Measurements ==="
curl -s "$BASE_URL/measurements?limit=10" | jq

# Get 24h statistics
START=$(date -u -d '24 hours ago' +%Y-%m-%dT%H:%M:%SZ)
END=$(date -u +%Y-%m-%dT%H:%M:%SZ)
echo -e "\n=== 24h Statistics ==="
curl -s "$BASE_URL/measurements/stats?start=$START&end=$END" | jq

# Get sensor info
echo -e "\n=== Sensor Info ==="
curl -s "$BASE_URL/sensors" | jq
```

## 📚 Documentation

For detailed information about the project:

- **[Architecture Documentation](docs/ARCHITECTURE.md)**: Complete architectural overview, design patterns, database schema, testing strategy, and migration guide
- **[Environment Variables](docs/ENV_VARS.md)**: Comprehensive configuration reference with examples for development, production, and Docker deployments
- **[Documentation Index](docs/README.md)**: Quick start guide and documentation overview

## 📄 License

This project is licensed under the [MIT](LICENSE) license.

## ⚠️ Disclaimer

This tool is provided for informational and personal use only.
It is not a certified medical device and should not be used to make health-related decisions.
Use it at your own risk.
