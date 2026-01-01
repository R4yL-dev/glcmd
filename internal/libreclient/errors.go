package libreclient

import "fmt"

// NetworkError represents a network-level error (connection failed, timeout, etc.)
type NetworkError struct {
	Err error
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error: %v", e.Err)
}

func (e *NetworkError) Unwrap() error {
	return e.Err
}

// AuthError represents an authentication error (401 Unauthorized)
type AuthError struct {
	StatusCode int
	Body       []byte
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("authentication failed: HTTP %d", e.StatusCode)
}

// RateLimitError represents a rate limit error (429 Too Many Requests)
type RateLimitError struct {
	StatusCode int
	Body       []byte
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded: HTTP %d", e.StatusCode)
}

// ServerError represents a server-side error (5xx)
type ServerError struct {
	StatusCode int
	Body       []byte
}

func (e *ServerError) Error() string {
	return fmt.Sprintf("server error: HTTP %d", e.StatusCode)
}

// HTTPError represents any other HTTP error
type HTTPError struct {
	StatusCode int
	Body       []byte
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP error: %d", e.StatusCode)
}
