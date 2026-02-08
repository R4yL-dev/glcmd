# Environment Variables

**Version**: 0.7.1
**Updated**: 2026-02-08

## Overview

glcmd is configured via environment variables for flexibility across different deployment environments (development, production, containers). All variables have sensible defaults.

The daemon (`glcore`) uses authentication, daemon, and database variables. The CLI client (`glcli`) uses only `GLCMD_API_URL`.

## Authentication Configuration

### GLCMD_EMAIL
- **Description**: LibreView follower account email
- **Required**: **Yes**
- **Example**: `GLCMD_EMAIL=follower@example.com`
- **Note**: Must be a LibreLinkUp follower account, not the primary patient account

---

### GLCMD_PASSWORD
- **Description**: LibreView follower account password
- **Required**: **Yes**
- **Example**: `GLCMD_PASSWORD=your_secure_password`
- **Security**: Use strong passwords. Consider using secrets management in production.

---

## Daemon Configuration

### GLCMD_FETCH_INTERVAL
- **Description**: How often to fetch new glucose data from LibreView API
- **Format**: Go duration format (e.g., `5m`, `1h`, `90s`)
- **Default**: `5m` (5 minutes)
- **Example**: `GLCMD_FETCH_INTERVAL=3m`
- **Note**: Minimum recommended: 1 minute (to avoid API rate limiting)

**Usage**:
```bash
# Default: 5 minutes
GLCMD_FETCH_INTERVAL=5m

# More frequent: 2 minutes
GLCMD_FETCH_INTERVAL=2m

# Less frequent: 10 minutes
GLCMD_FETCH_INTERVAL=10m
```

---

### GLCMD_API_PORT
- **Description**: HTTP port for unified API server (health, metrics, glucose endpoints)
- **Default**: `8080`
- **Example**: `GLCMD_API_PORT=9090`
- **Used by**: `glcore`
- **Endpoints**:
  - `GET /health` - Health status (returns 200 if healthy, 503 if degraded/unhealthy)
  - `GET /metrics` - Runtime metrics (memory, goroutines, uptime)
  - `GET /v1/measurements/latest` - Latest glucose measurement
  - `GET /v1/measurements` - Paginated measurements
  - `GET /v1/measurements/stats` - Statistics with time-in-range
  - `GET /v1/sensors` - Current sensor information
  - `GET /v1/sensors/history` - Sensor history
  - `GET /v1/sensors/stats` - Sensor lifecycle statistics

**Usage**:
```bash
# Default port
GLCMD_API_PORT=8080

# Custom port
GLCMD_API_PORT=9090

# Check health
curl http://localhost:8080/health
curl http://localhost:8080/metrics
```

---

### GLCMD_API_URL
- **Description**: Base URL for the glcore API server
- **Default**: `http://localhost:8080`
- **Example**: `GLCMD_API_URL=http://192.168.1.100:8080`
- **Used by**: `glcli`
- **Note**: Can also be set per-command with the `--api-url` flag

**Usage**:
```bash
# Default (local daemon)
GLCMD_API_URL=http://localhost:8080

# Remote daemon
GLCMD_API_URL=http://remote-server:8080

# Or use the flag
glcli --api-url http://remote-server:8080 stats
```

---

## Logging Configuration

### GLCMD_LOG_FORMAT
- **Description**: Output format for console logging
- **Values**: `text` | `json`
- **Default**: `text`
- **Example**: `GLCMD_LOG_FORMAT=json`
- **Note**: JSON format is useful for log aggregation systems (ELK, Loki, etc.)

**Usage**:
```bash
# Human-readable text output (default)
GLCMD_LOG_FORMAT=text
# Output: time=2026-02-07T10:30:00Z level=INFO msg="authenticated" duration=1.2s

# JSON output for log aggregation
GLCMD_LOG_FORMAT=json
# Output: {"time":"2026-02-07T10:30:00Z","level":"INFO","msg":"authenticated","duration":"1.2s"}
```

---

### GLCMD_LOG_LEVEL
- **Description**: Minimum log level to output
- **Values**: `debug` | `info` | `warn` | `error`
- **Default**: `info`
- **Example**: `GLCMD_LOG_LEVEL=debug`
- **Note**: Logs are output to stderr

**Log Levels**:
- `debug`: Verbose output including API requests and internal operations
- `info`: Standard operational messages (startup, fetch cycles, measurements)
- `warn`: Warning conditions that may require attention
- `error`: Error conditions only

**Usage**:
```bash
# Production (recommended)
GLCMD_LOG_LEVEL=info

# Development/debugging
GLCMD_LOG_LEVEL=debug

# Minimal output
GLCMD_LOG_LEVEL=warn
```

---

## Database Configuration

### GLCMD_DB_PATH
- **Description**: Path to SQLite database file
- **Default**: `./data/glcmd.db`
- **Example**: `GLCMD_DB_PATH=/var/lib/glcmd/glcmd.db`
- **Note**: Directory must exist before starting the application

**Usage**:
```bash
# Default location
GLCMD_DB_PATH=./data/glcmd.db

# Custom location
GLCMD_DB_PATH=/home/user/.glcmd/database.db
```

**Directory Creation**:
```bash
mkdir -p ./data  # Create directory if it doesn't exist
```

---

### GLCMD_DB_MAX_OPEN_CONNS
- **Description**: Maximum number of open database connections
- **Default**: `1`
- **Example**: `GLCMD_DB_MAX_OPEN_CONNS=1`
- **Note**: SQLite should always use 1 (single writer limitation)

---

### GLCMD_DB_MAX_IDLE_CONNS
- **Description**: Maximum number of idle connections in the pool
- **Default**: `1`
- **Example**: `GLCMD_DB_MAX_IDLE_CONNS=1`
- **Recommendation**: Set to same as MAX_OPEN_CONNS for consistent pool size

---

### GLCMD_DB_LOG_LEVEL
- **Description**: GORM logging level
- **Values**: `silent` | `error` | `warn` | `info`
- **Default**: `warn`
- **Example**: `GLCMD_DB_LOG_LEVEL=info`

**Log Levels**:
- `silent`: No logging (production)
- `error`: Only errors (production)
- `warn`: Errors + slow queries (recommended for production)
- `info`: All queries (development/debugging)

**Usage**:
```bash
# Production
GLCMD_DB_LOG_LEVEL=warn

# Development/debugging
GLCMD_DB_LOG_LEVEL=info
```

---

## Configuration Examples

### Development

```bash
# Authentication
GLCMD_EMAIL=dev@example.com
GLCMD_PASSWORD=dev_password

# Database
GLCMD_DB_PATH=./data/glcmd.db
GLCMD_DB_LOG_LEVEL=info

# Logging
GLCMD_LOG_LEVEL=debug
GLCMD_LOG_FORMAT=text

# Application
GLCMD_FETCH_INTERVAL=5m
GLCMD_API_PORT=8080
```

### Production

```bash
# Authentication
GLCMD_EMAIL=user@example.com
GLCMD_PASSWORD=secure_password

# Database
GLCMD_DB_PATH=/var/lib/glcmd/glcmd.db
GLCMD_DB_LOG_LEVEL=warn
GLCMD_DB_MAX_OPEN_CONNS=1
GLCMD_DB_MAX_IDLE_CONNS=1

# Logging
GLCMD_LOG_LEVEL=info
GLCMD_LOG_FORMAT=text

# Application
GLCMD_FETCH_INTERVAL=5m
GLCMD_API_PORT=8080
```

### Docker Compose Example

```yaml
version: '3.8'

services:
  glcore:
    build: .
    environment:
      # Authentication
      GLCMD_EMAIL: ${LIBREVIEW_EMAIL}
      GLCMD_PASSWORD: ${LIBREVIEW_PASSWORD}

      # Database
      GLCMD_DB_PATH: /data/glcmd.db
      GLCMD_DB_LOG_LEVEL: warn

      # Logging
      GLCMD_LOG_LEVEL: info
      GLCMD_LOG_FORMAT: json

      # Application
      GLCMD_FETCH_INTERVAL: 5m
      GLCMD_API_PORT: 8080
    volumes:
      - glcmd_data:/data
    ports:
      - "8080:8080"
    restart: unless-stopped

volumes:
  glcmd_data:
```

## Environment Variable Loading

### Order of Precedence

1. **Environment variables** (highest priority)
2. **Default values** in code (lowest priority)

### Loading via .env File

While glcmd doesn't natively support `.env` files, you can load them before starting:

```bash
# Install godotenv (optional)
go install github.com/joho/godotenv/cmd/godotenv@latest

# Run with .env
godotenv -f .env ./bin/glcore
```

**Example .env file**:
```env
GLCMD_EMAIL=user@example.com
GLCMD_PASSWORD=password
GLCMD_DB_PATH=./data/glcmd.db
GLCMD_DB_LOG_LEVEL=info
GLCMD_LOG_LEVEL=info
GLCMD_LOG_FORMAT=text
GLCMD_FETCH_INTERVAL=5m
GLCMD_API_PORT=8080
```

### Systemd Service Example

```ini
[Unit]
Description=glcmd glucose monitoring daemon
After=network.target

[Service]
Type=simple
User=glcmd
Group=glcmd
WorkingDirectory=/opt/glcmd

# Environment variables
Environment="GLCMD_EMAIL=user@example.com"
Environment="GLCMD_PASSWORD=password"
Environment="GLCMD_DB_PATH=/var/lib/glcmd/glcmd.db"
Environment="GLCMD_DB_LOG_LEVEL=warn"
Environment="GLCMD_LOG_LEVEL=info"
Environment="GLCMD_LOG_FORMAT=text"
Environment="GLCMD_FETCH_INTERVAL=5m"
Environment="GLCMD_API_PORT=8080"

ExecStart=/opt/glcmd/glcore
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

## Security Best Practices

### Sensitive Variables

The `GLCMD_PASSWORD` variable contains sensitive information.

**Recommendations**:
1. **Never commit** to version control
2. **Use secrets management** (Vault, AWS Secrets Manager, etc.)
3. **Restrict file permissions** if using .env files (`chmod 600 .env`)
4. **Use environment-specific values** (different passwords per environment)

### Example: Docker Secrets

```yaml
version: '3.8'

services:
  glcore:
    image: glcmd:latest
    environment:
      GLCMD_EMAIL: user@example.com
      GLCMD_DB_PATH: /data/glcmd.db
      GLCMD_LOG_LEVEL: info
      # Password from secret
      GLCMD_PASSWORD_FILE: /run/secrets/libreview_password
    secrets:
      - libreview_password

secrets:
  libreview_password:
    external: true
```

## Validation

The application validates configuration on startup and will exit with an error if:
- Database connection fails
- Invalid duration formats
- Invalid enum values (GLCMD_DB_LOG_LEVEL, GLCMD_LOG_LEVEL, etc.)
- Missing required credentials (email/password)

**Example Error Messages**:
```
Error: failed to connect to database: unable to open database file
Error: invalid GLCMD_LOG_LEVEL value: 'verbose' (must be 'debug', 'info', 'warn', or 'error')
```

## Troubleshooting

### Database Connection Issues

**Problem**: `failed to connect to database`

**Solutions**:
1. Check `GLCMD_DB_PATH` directory exists
2. Verify file permissions on database file
3. Check database logs for connection errors

### Lock/Busy Errors

**Problem**: `database is locked` or `SQLITE_BUSY`

**Solutions**:
1. Ensure `GLCMD_DB_MAX_OPEN_CONNS=1` for SQLite
2. Check for long-running queries blocking writes
3. Ensure only one glcore instance is running

### Performance Issues

**Problem**: Slow database operations

**Solutions**:
1. Set `GLCMD_DB_LOG_LEVEL=info` to see slow queries
2. Check WAL mode enabled (SQLite)

## Upgrading from v0.4.x to v0.5.0

### Removed Environment Variables

The following environment variables have been removed:

- `GLCMD_DISPLAY_INTERVAL` - Periodic display feature removed
- `GLCMD_ENABLE_EMOJIS` - Emoji display feature removed

If these variables are set, they will be ignored.

### New Environment Variables

- `GLCMD_LOG_FORMAT`: Log output format (`text` or `json`)
- `GLCMD_LOG_LEVEL`: Log verbosity (`debug`, `info`, `warn`, `error`)

### Behavioral Changes

- The daemon no longer displays periodic glucose readings to the console
- All glucose data is accessed via the REST API or CLI client (`glcli`)
- Logs are output to stderr instead of files

---

## Reference: Default Values Summary

| Variable | Default | Type |
|----------|---------|------|
| GLCMD_EMAIL | (required) | string |
| GLCMD_PASSWORD | (required) | string |
| GLCMD_FETCH_INTERVAL | `5m` | duration |
| GLCMD_API_PORT | `8080` | int |
| GLCMD_API_URL | `http://localhost:8080` | string |
| GLCMD_LOG_FORMAT | `text` | string |
| GLCMD_LOG_LEVEL | `info` | string |
| GLCMD_DB_PATH | `./data/glcmd.db` | string |
| GLCMD_DB_MAX_OPEN_CONNS | `1` | int |
| GLCMD_DB_MAX_IDLE_CONNS | `1` | int |
| GLCMD_DB_LOG_LEVEL | `warn` | string |
