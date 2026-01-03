package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/R4yL-dev/glcmd/internal/persistence"
)

// ErrorResponse represents a standard error response structure
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains the error code and message
type ErrorDetail struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// NewValidationError creates a new validation error
func NewValidationError(message string) error {
	return &ValidationError{Message: message}
}

// isValidationError checks if an error is a validation error
func isValidationError(err error) bool {
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}

// handleError handles errors and writes appropriate HTTP responses
func handleError(w http.ResponseWriter, err error, logger *slog.Logger) {
	var statusCode int
	var message string

	switch {
	case errors.Is(err, persistence.ErrNotFound):
		statusCode = http.StatusNotFound
		message = "Resource not found"
	case errors.Is(err, context.DeadlineExceeded):
		statusCode = http.StatusGatewayTimeout
		message = "Request timeout"
	case isValidationError(err):
		statusCode = http.StatusBadRequest
		message = err.Error()
	default:
		statusCode = http.StatusInternalServerError
		message = "Internal server error"
		logger.Error("unhandled error", "error", err)
	}

	writeJSONError(w, statusCode, message)
}

// writeJSONError writes a JSON error response
func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error: ErrorDetail{
			Code:    statusCode,
			Message: message,
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// If encoding fails, we can't do much - log it
		slog.Error("failed to encode error response", "error", err)
	}
}
