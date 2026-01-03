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

## 🌐 HTTP API

When running in daemon mode, `glcmd` exposes a REST API on port 8080 for programmatic access to glucose data, health checks, and metrics.

**Quick Examples:**
```bash
# Health check
curl http://localhost:8080/health | jq

# Get latest glucose measurement
curl http://localhost:8080/measurements/latest | jq

# Get statistics for last 7 days
START=$(date -u -d '7 days ago' +%Y-%m-%dT%H:%M:%SZ)
END=$(date -u +%Y-%m-%dT%H:%M:%SZ)
curl "http://localhost:8080/measurements/stats?start=$START&end=$END" | jq
```

**Available Endpoints:**
- `GET /health` - Daemon health status
- `GET /metrics` - Runtime metrics (uptime, memory, goroutines)
- `GET /measurements/latest` - Most recent glucose reading
- `GET /measurements` - Paginated measurements with filters
- `GET /measurements/stats` - Statistics with time-in-range analysis
- `GET /sensors` - Sensor information

For complete API documentation with all parameters, response formats, and examples, see **[API Reference](docs/API.md)**.

## 📚 Documentation

For detailed information about the project:

- **[API Reference](docs/API.md)**: Complete HTTP API documentation with all endpoints, parameters, response formats, and usage examples
- **[Architecture Documentation](docs/ARCHITECTURE.md)**: Complete architectural overview, design patterns, database schema, testing strategy, and migration guide
- **[Environment Variables](docs/ENV_VARS.md)**: Comprehensive configuration reference with examples for development, production, and Docker deployments
- **[Documentation Index](docs/README.md)**: Quick start guide and documentation overview

## 📄 License

This project is licensed under the [MIT](LICENSE) license.

## ⚠️ Disclaimer

This tool is provided for informational and personal use only.
It is not a certified medical device and should not be used to make health-related decisions.
Use it at your own risk.
