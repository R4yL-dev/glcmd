# Timezone Handling Issue

## Problem

LibreView API returns two timestamps for glucose measurements:

- **`FactoryTimestamp`**: Sensor time (appears to be UTC)
- **`Timestamp`**: Phone/device time (appears to be in user's local timezone)

### Example from real data:
```json
{
  "FactoryTimestamp": "1/1/2026 1:52:27 PM",  // 13:52 UTC
  "Timestamp": "1/1/2026 2:52:27 PM"           // 14:52 (UTC+1, Europe/Paris winter)
}
```

The 1-hour difference suggests `FactoryTimestamp` is UTC and `Timestamp` is local time (UTC+1 in this case).

## Current Behavior

Our `ParseLibreViewTimestamp()` function uses `time.Parse()` which:
1. Parses the string without timezone information
2. Returns a `time.Time` with location = **UTC** (Go's default for Parse without timezone)
3. **This is INCORRECT for `Timestamp` field** if it's in local timezone

### Impact

- **For users in UTC timezone**: No problem
- **For users in other timezones** (e.g., Europe/Paris UTC+1):
  - All timestamps are off by their timezone offset
  - Example: A measurement at 14:00 local time is stored as 14:00 UTC instead of 13:00 UTC

### Why This Matters

1. **Statistics calculations**: Time-based queries (e.g., "last 24 hours") may include wrong data
2. **Display**: If we later display timestamps in user's timezone, they'll be wrong
3. **Comparisons**: Comparing measurements across days may give incorrect results

## Proposed Solutions

### Option 1: Add Timezone Configuration (RECOMMENDED)

Add a `TIMEZONE` environment variable or config setting:

```go
// In config package
var UserTimezone = os.Getenv("GL_TIMEZONE") // e.g., "Europe/Paris"

// In timeparser
func ParseLibreViewTimestamp(timestamp string, isLocalTime bool) (time.Time, error) {
    t, err := time.Parse(libreViewLayout, timestamp)
    if err != nil {
        return time.Time{}, err
    }

    if isLocalTime && UserTimezone != "" {
        loc, err := time.LoadLocation(UserTimezone)
        if err == nil {
            // Interpret timestamp as being in user's timezone
            t = time.Date(t.Year(), t.Month(), t.Day(),
                         t.Hour(), t.Minute(), t.Second(),
                         t.Nanosecond(), loc)
        }
    }

    return t.UTC(), nil
}
```

Usage:
```go
// For FactoryTimestamp (UTC)
factoryTime, _ := ParseLibreViewTimestamp(data.FactoryTimestamp, false)

// For Timestamp (local time)
timestamp, _ := ParseLibreViewTimestamp(data.Timestamp, true)
```

### Option 2: Auto-detect from FactoryTimestamp difference

Calculate offset: `Timestamp - FactoryTimestamp` and apply it:

```go
offset := timestamp.Sub(factoryTimestamp)
correctedTimestamp := timestamp.Add(-offset)
```

**Problem**: Assumes FactoryTimestamp and Timestamp always have the same offset, which may not be true.

### Option 3: Leave as-is (CURRENT)

- Document the limitation
- Works fine for users in UTC
- Works "okay" for statistics (time difference calculations are still correct)
- Only affects absolute time values

## Decision

**For now**: Leave as-is (Option 3), with this documentation.

**Future**: Implement Option 1 when we add configuration system.

## Testing Note

To test with real data, check your dumps:
```bash
# Check the time difference
grep -A1 "FactoryTimestamp" exploration/dumps/connections.json
```

If the difference is exactly your timezone offset, this confirms the issue.
