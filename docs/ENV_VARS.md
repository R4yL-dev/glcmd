# Environment Variables

**Version**: 0.2.0
**Updated**: 2026-01-03

## Overview

glcmd is configured via environment variables for flexibility across different deployment environments (development, production, containers). All variables have sensible defaults.

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

### GLCMD_DISPLAY_INTERVAL
- **Description**: How often to display the latest glucose measurement in logs
- **Format**: Go duration format (e.g., `1m`, `30s`, `2m`)
- **Default**: `1m` (1 minute)
- **Example**: `GLCMD_DISPLAY_INTERVAL=30s`

**Usage**:
```bash
# Default: 1 minute
GLCMD_DISPLAY_INTERVAL=1m

# More frequent: 30 seconds
GLCMD_DISPLAY_INTERVAL=30s

# Less frequent: 5 minutes
GLCMD_DISPLAY_INTERVAL=5m
```

---

### GLCMD_ENABLE_EMOJIS
- **Description**: Enable emoji display in trend arrows and status indicators
- **Values**: `true` | `false`
- **Default**: `true`
- **Example**: `GLCMD_ENABLE_EMOJIS=false`
- **Note**: Set to `false` for legacy terminals or log parsing systems

**Usage**:
```bash
# Enable emojis (default)
GLCMD_ENABLE_EMOJIS=true
# Output: ⬆️⬆️ 🟢 Normal

# Disable emojis (ASCII only)
GLCMD_ENABLE_EMOJIS=false
# Output: Rising rapidly Normal
```

---

### GLCMD_API_PORT
- **Description**: HTTP port for unified API server (health, metrics, glucose endpoints)
- **Default**: `8080`
- **Example**: `GLCMD_API_PORT=9090`
- **Endpoints**:
  - `GET /health` - Health status (returns 200 if healthy, 503 if degraded/unhealthy)
  - `GET /metrics` - Runtime metrics (memory, goroutines, uptime)
  - `GET /measurements/latest` - Latest glucose measurement
  - `GET /measurements` - Paginated measurements
  - `GET /measurements/stats` - Statistics with time-in-range
  - `GET /sensors` - Sensor information

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

## Database Configuration

### GLCMD_DB_TYPE
- **Description**: Database type to use
- **Values**: `sqlite` | `postgres`
- **Default**: `sqlite`
- **Example**: `GLCMD_DB_TYPE=sqlite`

**Usage**:
```bash
# SQLite (default)
GLCMD_DB_TYPE=sqlite

# PostgreSQL (future containerized deployment)
GLCMD_DB_TYPE=postgres
```

---

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

### GLCMD_DB_HOST
- **Description**: PostgreSQL server hostname or IP address
- **Default**: `localhost`
- **Example**: `GLCMD_DB_HOST=postgres.example.com`
- **Applies to**: `GLCMD_DB_TYPE=postgres` only

---

### GLCMD_DB_PORT
- **Description**: PostgreSQL server port
- **Default**: `5432`
- **Example**: `GLCMD_DB_PORT=5432`
- **Applies to**: `GLCMD_DB_TYPE=postgres` only

---

### GLCMD_DB_USER
- **Description**: PostgreSQL username
- **Default**: `postgres`
- **Example**: `GLCMD_DB_USER=glcmd_user`
- **Applies to**: `GLCMD_DB_TYPE=postgres` only

---

### GLCMD_DB_PASSWORD
- **Description**: PostgreSQL password
- **Default**: `postgres`
- **Example**: `GLCMD_DB_PASSWORD=secure_password_here`
- **Applies to**: `GLCMD_DB_TYPE=postgres` only
- **Security**: Use strong passwords in production. Consider secrets management.

---

### GLCMD_DB_DBNAME
- **Description**: PostgreSQL database name
- **Default**: `glcmd`
- **Example**: `GLCMD_DB_DBNAME=glucose_monitoring`
- **Applies to**: `GLCMD_DB_TYPE=postgres` only
- **Note**: Database must be created before starting the application

---

### GLCMD_DB_SSLMODE
- **Description**: PostgreSQL SSL mode
- **Values**: `disable` | `require` | `verify-ca` | `verify-full`
- **Default**: `disable`
- **Example**: `GLCMD_DB_SSLMODE=require`
- **Applies to**: `GLCMD_DB_TYPE=postgres` only
- **Production**: Use `require` or higher for production deployments

---

### GLCMD_DB_MAX_OPEN_CONNS
- **Description**: Maximum number of open database connections
- **Default**: `1` (SQLite), `25` (PostgreSQL)
- **Example**: `GLCMD_DB_MAX_OPEN_CONNS=10`
- **Note**: SQLite should always use 1 (single writer limitation)

**Usage**:
```bash
# SQLite (always use 1)
GLCMD_DB_MAX_OPEN_CONNS=1

# PostgreSQL (tune based on load)
GLCMD_DB_MAX_OPEN_CONNS=25
```

---

### GLCMD_DB_MAX_IDLE_CONNS
- **Description**: Maximum number of idle connections in the pool
- **Default**: `1` (SQLite), `5` (PostgreSQL)
- **Example**: `GLCMD_DB_MAX_IDLE_CONNS=5`
- **Recommendation**: Set to same as MAX_OPEN_CONNS for consistent pool size

---

### DB_CONN_MAX_LIFETIME
- **Description**: Maximum lifetime of a connection (duration string)
- **Default**: `1h`
- **Example**: `DB_CONN_MAX_LIFETIME=30m`
- **Format**: Valid Go duration string (`10s`, `5m`, `1h`, `24h`)

**Valid Duration Formats**:
- `10s` - 10 seconds
- `5m` - 5 minutes
- `1h` - 1 hour
- `24h` - 24 hours

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

## Retry Configuration

### DB_RETRY_MAX_RETRIES
- **Description**: Maximum number of retry attempts for database operations
- **Default**: `3`
- **Example**: `DB_RETRY_MAX_RETRIES=5`
- **Range**: `0` (no retries) to `10` (aggressive retrying)

---

### DB_RETRY_INITIAL_BACKOFF
- **Description**: Initial backoff duration before first retry
- **Default**: `100ms`
- **Example**: `DB_RETRY_INITIAL_BACKOFF=50ms`
- **Format**: Valid Go duration string

---

### DB_RETRY_MAX_BACKOFF
- **Description**: Maximum backoff duration (cap for exponential backoff)
- **Default**: `500ms`
- **Example**: `DB_RETRY_MAX_BACKOFF=1s`
- **Format**: Valid Go duration string

---

### DB_RETRY_MULTIPLIER
- **Description**: Backoff multiplier for exponential backoff
- **Default**: `2.0`
- **Example**: `DB_RETRY_MULTIPLIER=1.5`
- **Range**: `1.0` (linear) to `3.0` (aggressive exponential)

**Backoff Progression Example** (default config):
- 1st retry: 100ms
- 2nd retry: 200ms (100ms × 2.0)
- 3rd retry: 400ms (200ms × 2.0)

---

## Configuration Examples

### Development (SQLite)

```bash
# Database
GLCMD_DB_TYPE=sqlite
GLCMD_DB_PATH=./data/glcmd.db
GLCMD_DB_LOG_LEVEL=info

# Retry (more aggressive for debugging)
DB_RETRY_MAX_RETRIES=5
DB_RETRY_INITIAL_BACKOFF=50ms

# Application
GLCMD_EMAIL=dev@example.com
GLCMD_PASSWORD=dev_password
GLCMD_FETCH_INTERVAL=5m
GLCMD_API_PORT=8080
```

### Production (SQLite)

```bash
# Database
GLCMD_DB_TYPE=sqlite
GLCMD_DB_PATH=/var/lib/glcmd/glcmd.db
GLCMD_DB_LOG_LEVEL=warn
GLCMD_DB_MAX_OPEN_CONNS=1
GLCMD_DB_MAX_IDLE_CONNS=1

# Retry (standard)
DB_RETRY_MAX_RETRIES=3
DB_RETRY_INITIAL_BACKOFF=100ms
DB_RETRY_MAX_BACKOFF=500ms

# Application
GLCMD_EMAIL=user@example.com
GLCMD_PASSWORD=secure_password
GLCMD_FETCH_INTERVAL=5m
GLCMD_API_PORT=8080
```

### Production (PostgreSQL - Future)

```bash
# Database
GLCMD_DB_TYPE=postgres
GLCMD_DB_HOST=postgres-service
GLCMD_DB_PORT=5432
GLCMD_DB_USER=glcmd_app
GLCMD_DB_PASSWORD=strong_password_here
GLCMD_DB_DBNAME=glcmd_production
GLCMD_DB_SSLMODE=require
GLCMD_DB_LOG_LEVEL=warn

# Connection pooling
GLCMD_DB_MAX_OPEN_CONNS=25
GLCMD_DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=1h

# Retry
DB_RETRY_MAX_RETRIES=3
DB_RETRY_INITIAL_BACKOFF=100ms
DB_RETRY_MAX_BACKOFF=500ms
DB_RETRY_MULTIPLIER=2.0

# Application
GLCMD_EMAIL=user@example.com
GLCMD_PASSWORD=secure_password
GLCMD_FETCH_INTERVAL=5m
GLCMD_API_PORT=8080
```

### Docker Compose Example

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: glcmd_app
      POSTGRES_PASSWORD: secure_password
      POSTGRES_DB: glcmd_production
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  glcmd:
    build: .
    depends_on:
      - postgres
    environment:
      # Database
      GLCMD_DB_TYPE: postgres
      GLCMD_DB_HOST: postgres
      GLCMD_DB_PORT: 5432
      GLCMD_DB_USER: glcmd_app
      GLCMD_DB_PASSWORD: secure_password
      GLCMD_DB_DBNAME: glcmd_production
      GLCMD_DB_SSLMODE: disable  # Internal network
      GLCMD_DB_LOG_LEVEL: warn

      # Connection pooling
      GLCMD_DB_MAX_OPEN_CONNS: 25
      GLCMD_DB_MAX_IDLE_CONNS: 5
      DB_CONN_MAX_LIFETIME: 1h

      # Application
      GLCMD_EMAIL: ${LIBREVIEW_EMAIL}
      GLCMD_PASSWORD: ${LIBREVIEW_PASSWORD}
      GLCMD_FETCH_INTERVAL: 5m
      GLCMD_API_PORT: 8080
    ports:
      - "8080:8080"
    restart: unless-stopped

volumes:
  postgres_data:
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
godotenv -f .env ./glcmd daemon
```

**Example .env file**:
```env
GLCMD_DB_TYPE=sqlite
GLCMD_DB_PATH=./data/glcmd.db
GLCMD_DB_LOG_LEVEL=info
GLCMD_EMAIL=user@example.com
GLCMD_PASSWORD=password
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
Environment="GLCMD_DB_TYPE=sqlite"
Environment="GLCMD_DB_PATH=/var/lib/glcmd/glcmd.db"
Environment="GLCMD_DB_LOG_LEVEL=warn"
Environment="GLCMD_EMAIL=user@example.com"
Environment="GLCMD_PASSWORD=password"
Environment="GLCMD_FETCH_INTERVAL=5m"
Environment="GLCMD_API_PORT=8080"

ExecStart=/opt/glcmd/glcmd
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

## Security Best Practices

### Sensitive Variables

These variables contain sensitive information:
- `GLCMD_PASSWORD`
- `GLCMD_DB_PASSWORD`

**Recommendations**:
1. **Never commit** to version control
2. **Use secrets management** (Vault, AWS Secrets Manager, etc.)
3. **Restrict file permissions** if using .env files (`chmod 600 .env`)
4. **Use environment-specific values** (different passwords per environment)

### Example: Docker Secrets

```yaml
version: '3.8'

services:
  glcmd:
    image: glcmd:latest
    environment:
      GLCMD_DB_TYPE: postgres
      GLCMD_DB_HOST: postgres
      GLCMD_DB_USER: glcmd_app
      GLCMD_DB_DBNAME: glcmd_production
      # Passwords from secrets
      GLCMD_PASSWORD_FILE: /run/secrets/libreview_password
      GLCMD_DB_PASSWORD_FILE: /run/secrets/db_password
    secrets:
      - libreview_password
      - db_password

secrets:
  libreview_password:
    external: true
  db_password:
    external: true
```

## Validation

The application validates configuration on startup and will exit with an error if:
- Database connection fails
- Invalid duration formats
- Invalid enum values (GLCMD_DB_TYPE, GLCMD_DB_LOG_LEVEL, etc.)
- Missing required credentials (email/password)

**Example Error Messages**:
```
Error: invalid GLCMD_DB_TYPE value: 'mysql' (must be 'sqlite' or 'postgres')
Error: failed to connect to database: unable to open database file
Error: invalid duration format for DB_CONN_MAX_LIFETIME: 'invalid'
```

## Troubleshooting

### Database Connection Issues

**Problem**: `failed to connect to database`

**Solutions**:
1. Check `GLCMD_DB_PATH` directory exists
2. Verify file permissions on database file
3. For PostgreSQL: verify host, port, credentials
4. Check database logs for connection errors

### Lock/Busy Errors

**Problem**: `database is locked` or `SQLITE_BUSY`

**Solutions**:
1. Ensure `GLCMD_DB_MAX_OPEN_CONNS=1` for SQLite
2. Increase `DB_RETRY_MAX_RETRIES`
3. Increase `DB_RETRY_MAX_BACKOFF`
4. Check for long-running queries blocking writes

### Performance Issues

**Problem**: Slow database operations

**Solutions**:
1. Set `GLCMD_DB_LOG_LEVEL=info` to see slow queries
2. For PostgreSQL: tune connection pool (`GLCMD_DB_MAX_OPEN_CONNS`)
3. Check WAL mode enabled (SQLite)
4. Monitor retry frequency in logs

## Migration from SQLite to PostgreSQL

### Step 1: Export SQLite Data

```bash
# Dump schema and data
sqlite3 ./data/glcmd.db .dump > backup.sql

# Or use a tool like pgloader
pgloader ./data/glcmd.db postgresql://user:pass@host/dbname
```

### Step 2: Update Environment Variables

```bash
# Change from SQLite
GLCMD_DB_TYPE=sqlite
GLCMD_DB_PATH=./data/glcmd.db

# To PostgreSQL
GLCMD_DB_TYPE=postgres
GLCMD_DB_HOST=localhost
GLCMD_DB_PORT=5432
GLCMD_DB_USER=glcmd_app
GLCMD_DB_PASSWORD=password
GLCMD_DB_DBNAME=glcmd_production
GLCMD_DB_SSLMODE=require
```

### Step 3: Restart Application

```bash
# Application will auto-migrate schema on PostgreSQL
./glcmd daemon
```

## Upgrading from v0.1.x to v0.2.0

### Environment Variable Changes

The following environment variables have been renamed for clarity and consistency:

- `GLCMD_HEALTHCHECK_PORT` → `GLCMD_API_PORT` (same functionality, renamed for clarity)
- `GLCMD_INTERVAL` → `GLCMD_FETCH_INTERVAL` (if using legacy variable)

All other variables remain compatible. Update your configuration files and deployment scripts accordingly.

### API Changes

- All API endpoints now unified on single port (default: 8080)
- No breaking changes to endpoint specifications
- See [API Documentation](API.md) for complete current specification
- See [CHANGELOG](../CHANGELOG.md) for detailed list of all changes

---

## Reference: Default Values Summary

| Variable | Default | Type |
|----------|---------|------|
| GLCMD_DB_TYPE | `sqlite` | string |
| GLCMD_DB_PATH | `./data/glcmd.db` | string |
| GLCMD_DB_HOST | `localhost` | string |
| GLCMD_DB_PORT | `5432` | int |
| GLCMD_DB_USER | `postgres` | string |
| GLCMD_DB_PASSWORD | `postgres` | string |
| GLCMD_DB_DBNAME | `glcmd` | string |
| GLCMD_DB_SSLMODE | `disable` | string |
| GLCMD_DB_MAX_OPEN_CONNS | `1` (SQLite), `25` (PG) | int |
| GLCMD_DB_MAX_IDLE_CONNS | `1` (SQLite), `5` (PG) | int |
| DB_CONN_MAX_LIFETIME | `1h` | duration |
| GLCMD_DB_LOG_LEVEL | `warn` | string |
| DB_RETRY_MAX_RETRIES | `3` | int |
| DB_RETRY_INITIAL_BACKOFF | `100ms` | duration |
| DB_RETRY_MAX_BACKOFF | `500ms` | duration |
| DB_RETRY_MULTIPLIER | `2.0` | float |
| GLCMD_FETCH_INTERVAL | `5m` | duration |
| GLCMD_DISPLAY_INTERVAL | `1m` | duration |
| GLCMD_ENABLE_EMOJIS | `true` | boolean |
| GLCMD_API_PORT | `8080` | int |
