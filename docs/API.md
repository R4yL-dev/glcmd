# API Documentation

When running in daemon mode, `glcmd` provides a unified HTTP API server on port 8080 (configurable via `GLCMD_API_PORT`). All endpoints return JSON responses with consistent formatting and pass through logging, recovery, and timeout middlewares.

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
    }
  }
}
```

**Example:**
```bash
curl http://localhost:8080/metrics | jq
```

---

### 3. Latest Measurement

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

### 4. Measurements List

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

### 5. Statistics

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

### 6. Sensors

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
