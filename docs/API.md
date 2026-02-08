# HTTP API Documentation

**Version**: 0.7.1
**Updated**: 2026-02-08
**Status**: Stable

`glcore` provides a unified HTTP API server on port 8080 (configurable via `GLCMD_API_PORT`). All endpoints return JSON responses with consistent formatting and pass through logging, recovery, and timeout middlewares. The SSE streaming endpoint is excluded from timeout middleware to support long-lived connections.

## API Stability

All endpoints in this documentation are stable and part of the 0.6.0 API contract. Breaking changes will be documented in [CHANGELOG.md](../CHANGELOG.md) and trigger a minor version bump.

## API Versioning

**Data endpoints** are versioned using a URL prefix for API stability. The current version is `/v1`.

**Versioned endpoints:**
- `/v1/glucose` - Paginated glucose measurements
- `/v1/glucose/latest` - Most recent glucose reading
- `/v1/glucose/stats` - Glucose statistics
- `/v1/sensor` - Paginated sensor list
- `/v1/sensor/latest` - Current active sensor
- `/v1/sensor/stats` - Sensor lifecycle statistics
- `/v1/stream` - Real-time event stream (SSE)

**Unversioned endpoints** (monitoring):
- `/health` - Health check
- `/metrics` - Runtime metrics

This versioning strategy allows future API evolution while maintaining backward compatibility.

## CORS Support

The API includes Cross-Origin Resource Sharing (CORS) headers to enable web frontend access:
- `Access-Control-Allow-Origin: *` - Allows all origins
- `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type, Authorization`
- `Access-Control-Max-Age: 3600` - Preflight cache duration

CORS preflight requests (`OPTIONS`) are handled automatically.

## Base URL

```
http://localhost:8080
```

## Response Format

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

## Endpoints

### 1. Health Check

**GET** `/health`

Returns the daemon and database health status.

**Response Codes:**
- `200 OK` - Service and database are healthy
- `503 Service Unavailable` - Service is degraded, unhealthy, or database is disconnected

**Response:**
```json
{
  "data": {
    "status": "healthy",
    "timestamp": "2025-01-03T10:30:00Z",
    "uptime": "2h15m30s",
    "consecutiveErrors": 0,
    "lastFetchError": "",
    "lastFetchTime": "2025-01-03T10:29:45Z",
    "databaseConnected": true,
    "dataFresh": true,
    "fetchInterval": "5m0s"
  }
}
```

**Status Values:**
- `healthy` - All systems operational, database connected, and data is fresh
- `degraded` - Some errors but still functional (1-2 consecutive errors), or data is stale
- `unhealthy` - Service experiencing issues (3+ consecutive errors) or database disconnected

**Database Status:**
- `databaseConnected: true` - Database is responsive
- `databaseConnected: false` - Database connection failed (returns 503)

**Data Freshness:**
- `dataFresh: true` - Last successful fetch was within 2x the configured fetch interval
- `dataFresh: false` - Data is stale (no successful fetch within 2x interval)
- `fetchInterval` - The configured polling interval (e.g., `"5m0s"`)
- When data becomes stale and status would otherwise be `healthy`, it degrades to `degraded`
- If `lastFetchTime` is zero (no fetch yet), data is considered fresh

**Example:**
```bash
curl http://localhost:8080/health | jq
```

---

### 2. Metrics

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
    },
    "sse": {
      "enabled": true,
      "subscribers": 2
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
```

**Field Descriptions:**
- `sse.enabled` - Whether the SSE event broker is active
- `sse.subscribers` - Number of currently connected SSE subscribers
- `database.openConnections` - Total number of open database connections
- `database.inUse` - Number of connections currently in use
- `database.idle` - Number of idle connections in the pool
- `database.waitCount` - Total number of connections waited for
- `database.waitDuration` - Total time blocked waiting for a new connection

**Example:**
```bash
curl http://localhost:8080/metrics | jq
```

---

### 3. Latest Glucose

**GET** `/v1/glucose/latest`

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
curl http://localhost:8080/v1/glucose/latest | jq
```

---

### 4. Glucose List

**GET** `/v1/glucose`

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
curl "http://localhost:8080/v1/glucose?limit=50" | jq

# Get measurements from last 24 hours
START=$(date -u -d '24 hours ago' +%Y-%m-%dT%H:%M:%SZ)
END=$(date -u +%Y-%m-%dT%H:%M:%SZ)
curl "http://localhost:8080/v1/glucose?start=$START&end=$END" | jq

# Get only warning/critical measurements
curl "http://localhost:8080/v1/glucose?color=2" | jq
curl "http://localhost:8080/v1/glucose?color=3" | jq

# Pagination example - get next page
curl "http://localhost:8080/v1/glucose?limit=100&offset=100" | jq
```

---

### 5. Glucose Statistics

**GET** `/v1/glucose/stats`

Returns glucose statistics for a specified time period, or all-time statistics if no date range is provided.

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `start` | string (RFC3339) | No | Start of time range (must be paired with `end`) |
| `end` | string (RFC3339) | No | End of time range (must be paired with `start`) |

If both `start` and `end` are omitted, returns all-time statistics. If provided, both must be specified together.

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
# Get all-time statistics
curl "http://localhost:8080/v1/glucose/stats" | jq

# Get statistics for last 7 days
START=$(date -u -d '7 days ago' +%Y-%m-%dT%H:%M:%SZ)
END=$(date -u +%Y-%m-%dT%H:%M:%SZ)
curl "http://localhost:8080/v1/glucose/stats?start=$START&end=$END" | jq

# Get statistics for a specific day
curl "http://localhost:8080/v1/glucose/stats?start=2025-01-01T00:00:00Z&end=2025-01-01T23:59:59Z" | jq

# Get statistics for last 30 days
START=$(date -u -d '30 days ago' +%Y-%m-%dT%H:%M:%SZ)
END=$(date -u +%Y-%m-%dT%H:%M:%SZ)
curl "http://localhost:8080/v1/glucose/stats?start=$START&end=$END" | jq
```

---

### 6. Latest Sensor

**GET** `/v1/sensor/latest`

Returns the current active sensor information with lifecycle details.

**Response:**
```json
{
  "data": {
    "serialNumber": "ABC123XYZ",
    "activation": "2025-12-28T18:02:35Z",
    "expiresAt": "2026-01-11T18:02:35Z",
    "endedAt": null,
    "lastMeasurementAt": "2026-01-05T10:30:00Z",
    "sensorType": 4,
    "durationDays": 14,
    "daysRemaining": 6.3,
    "daysElapsed": 7.7,
    "status": "running"
  }
}
```

**Field Descriptions:**
- `serialNumber` - Sensor serial number
- `activation` - Sensor activation timestamp
- `expiresAt` - Expected expiration timestamp
- `endedAt` - Actual end timestamp (null if still active)
- `lastMeasurementAt` - Timestamp of most recent measurement from this sensor (null if none)
- `sensorType` - Sensor type code
- `durationDays` - Expected sensor duration in days
- `daysRemaining` - Days remaining until expiration (running sensors only)
- `daysElapsed` - Days since activation (bounded by ExpiresAt for expired sensors)
- `actualDays` - Actual duration in days (stopped sensors with EndedAt only)
- `status` - Sensor status (`running`, `unresponsive`, `stopped`)

**Example:**
```bash
curl http://localhost:8080/v1/sensor/latest | jq
```

---

### 7. Sensor List

**GET** `/v1/sensor`

Returns a paginated list of all sensors with optional date filters.

**Query Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `limit` | integer | No | 100 | Number of results per page (1-1000) |
| `offset` | integer | No | 0 | Number of results to skip |
| `start` | string (RFC3339) | No | - | Filter sensors activated after this time |
| `end` | string (RFC3339) | No | - | Filter sensors activated before this time |

**Response:**
```json
{
  "data": [
    {
      "serialNumber": "ABC123XYZ",
      "activation": "2025-12-28T18:02:35Z",
      "expiresAt": "2026-01-11T18:02:35Z",
      "endedAt": "2026-01-10T14:22:00Z",
      "lastMeasurementAt": "2026-01-10T14:20:00Z",
      "sensorType": 4,
      "durationDays": 14,
      "daysElapsed": 12.8,
      "actualDays": 12.8,
      "status": "stopped"
    }
  ],
  "pagination": {
    "limit": 100,
    "offset": 0,
    "total": 5,
    "hasMore": false
  }
}
```

**Examples:**
```bash
# Get all sensors
curl "http://localhost:8080/v1/sensor" | jq

# Get sensors from last 6 months
START=$(date -u -d '6 months ago' +%Y-%m-%dT%H:%M:%SZ)
curl "http://localhost:8080/v1/sensor?start=$START" | jq
```

---

### 8. Sensor Statistics

**GET** `/v1/sensor/stats`

Returns aggregated sensor lifecycle statistics for a specified time period, or all-time statistics if no date range is provided.

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `start` | string (RFC3339) | No | Start of time range (filter sensors activated after this time) |
| `end` | string (RFC3339) | No | End of time range (filter sensors activated before this time) |

If both `start` and `end` are omitted, returns all-time statistics.

**Response:**
```json
{
  "data": {
    "period": {
      "start": "2025-01-01T00:00:00Z",
      "end": "2025-12-31T23:59:59Z"
    },
    "statistics": {
      "totalSensors": 5,
      "completedSensors": 4,
      "avgDuration": 13.2,
      "minDuration": 11.5,
      "maxDuration": 14.8,
      "avgExpected": 14.0,
      "avgDifference": -0.8
    },
    "current": {
      "serialNumber": "ABC123XYZ",
      "activation": "2025-12-28T18:02:35Z",
      "expiresAt": "2026-01-11T18:02:35Z",
      "sensorType": 4,
      "durationDays": 14,
      "daysRemaining": 6.3,
      "daysElapsed": 7.7,
      "status": "running"
    }
  }
}
```

**Field Descriptions:**
- `period` - Time period for the statistics (omitted for all-time queries)
- `statistics.totalSensors` - Total number of sensors tracked
- `statistics.completedSensors` - Number of completed sensors
- `statistics.avgDuration` - Average actual duration in days (ended sensors)
- `statistics.minDuration` - Shortest sensor duration in days
- `statistics.maxDuration` - Longest sensor duration in days
- `statistics.avgExpected` - Average expected duration in days
- `statistics.avgDifference` - Average difference between actual and expected duration (negative = ended early)
- `current` - Current active sensor information (null if none)

**Examples:**
```bash
# Get all-time sensor statistics
curl http://localhost:8080/v1/sensor/stats | jq

# Get sensor statistics for last 6 months
START=$(date -u -d '6 months ago' +%Y-%m-%dT%H:%M:%SZ)
END=$(date -u +%Y-%m-%dT%H:%M:%SZ)
curl "http://localhost:8080/v1/sensor/stats?start=$START&end=$END" | jq
```

---

### 9. Event Stream (SSE)

**GET** `/v1/stream`

Streams real-time events using Server-Sent Events (SSE). This endpoint maintains a long-lived connection and pushes events as they occur.

**Query Parameters:**

| Parameter | Type   | Required | Default | Description                              |
|-----------|--------|----------|---------|------------------------------------------|
| `types`   | string | No       | all     | Comma-separated event types to receive   |

**Event Types:**
- `glucose` - New glucose measurement
- `sensor` - Sensor status change (new sensor detected)
- `keepalive` - Heartbeat (every 30 seconds)

**Response Headers:**
- `Content-Type: text/event-stream`
- `Cache-Control: no-cache`
- `Connection: keep-alive`
- `X-Accel-Buffering: no` (disables nginx buffering)

**Event Format:**

Events are sent in standard SSE format:
```
event: glucose
data: {"timestamp":"2026-01-15T10:30:00Z","value":5.6,"valueInMgPerDl":101,"trendArrow":3,...}

event: sensor
data: {"serialNumber":"ABC123","activation":"2026-01-01T00:00:00Z","status":"running",...}

event: keepalive
data: {}
```

**Examples:**
```bash
# Stream all events
curl -N http://localhost:8080/v1/stream

# Stream glucose only
curl -N "http://localhost:8080/v1/stream?types=glucose"

# Stream multiple types
curl -N "http://localhost:8080/v1/stream?types=glucose,sensor"

# Using glcli
glcli watch                  # All events
glcli watch --only glucose   # Glucose only
glcli watch --json           # JSON output for scripting
```

**Notes:**
- The connection remains open until the client disconnects
- Keepalive events are sent every 30 seconds to detect dead connections
- Events are non-blocking: slow subscribers may miss events if their buffer fills up
- This endpoint bypasses the standard timeout middleware (5s) to allow long-lived connections

---

## Error Handling

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

## Complete Example Script

```bash
#!/bin/bash

BASE_URL="http://localhost:8080"

# Check health
echo "=== Health Check ==="
curl -s "$BASE_URL/health" | jq

# Get metrics
echo -e "\n=== Metrics ==="
curl -s "$BASE_URL/metrics" | jq

# Get latest glucose
echo -e "\n=== Latest Glucose ==="
curl -s "$BASE_URL/v1/glucose/latest" | jq

# Get last 10 measurements
echo -e "\n=== Last 10 Measurements ==="
curl -s "$BASE_URL/v1/glucose?limit=10" | jq

# Get all-time statistics
echo -e "\n=== All-Time Statistics ==="
curl -s "$BASE_URL/v1/glucose/stats" | jq

# Get 24h statistics
START=$(date -u -d '24 hours ago' +%Y-%m-%dT%H:%M:%SZ)
END=$(date -u +%Y-%m-%dT%H:%M:%SZ)
echo -e "\n=== 24h Statistics ==="
curl -s "$BASE_URL/v1/glucose/stats?start=$START&end=$END" | jq

# Get current sensor
echo -e "\n=== Current Sensor ==="
curl -s "$BASE_URL/v1/sensor/latest" | jq

# Get sensor history
echo -e "\n=== Sensor History ==="
curl -s "$BASE_URL/v1/sensor" | jq

# Get sensor stats
echo -e "\n=== Sensor Statistics ==="
curl -s "$BASE_URL/v1/sensor/stats" | jq
```
