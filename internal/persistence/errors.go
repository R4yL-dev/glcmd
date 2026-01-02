package persistence

import (
	"errors"
	"strings"
)

// Common database errors
var (
	ErrNotFound         = errors.New("record not found")
	ErrDuplicateKey     = errors.New("duplicate key violation")
	ErrConnectionFailed = errors.New("database connection failed")
	ErrTransactionFailed = errors.New("transaction failed")
)

// IsRetryable determines if an error should trigger retry logic.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()

	// SQLite-specific errors that are retryable
	if strings.Contains(errMsg, "database is locked") ||
		strings.Contains(errMsg, "SQLITE_BUSY") ||
		strings.Contains(errMsg, "database table is locked") {
		return true
	}

	// PostgreSQL-specific errors (future)
	if strings.Contains(errMsg, "connection refused") ||
		strings.Contains(errMsg, "connection reset") ||
		strings.Contains(errMsg, "timeout") {
		return true
	}

	// Generic connection errors
	if errors.Is(err, ErrConnectionFailed) {
		return true
	}

	return false
}
