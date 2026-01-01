// Package libreclient provides a clean, production-ready HTTP client
// for the LibreView API.
//
// Features:
//   - Context support for timeout/cancellation
//   - Proper error handling (distinguish 401, 429, 5xx, etc.)
//   - No unnecessary DNS lookups
//   - Simple, idiomatic Go
//   - Easy to test with custom http.Client
package libreclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// BaseURL is the LibreView API base URL (global endpoint)
	BaseURL = "https://api.libreview.io"

	// Default timeout for HTTP requests
	DefaultTimeout = 30 * time.Second
)

// Client is a LibreView API client.
type Client struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string
	version    string
	product    string
}

// NewClient creates a new LibreView API client.
//
// If httpClient is nil, a default client with 30s timeout is used.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: DefaultTimeout,
		}
	}

	return &Client{
		httpClient: httpClient,
		baseURL:    BaseURL,
		userAgent:  "Mozilla/5.0 (iPhone; CPU OS 17_4.1 like Mac OS X) AppleWebKit/536.26 (KHTML, like Gecko) Version/17.4.1 Mobile/10A5355d Safari/8536.25",
		version:    "4.16.0",
		product:    "llu.ios",
	}
}

// doRequest performs an HTTP request with proper error handling.
//
// It handles:
//   - Context cancellation/timeout
//   - HTTP status codes (returns appropriate error types)
//   - JSON decoding
//
// Optional auth can be provided via token and accountID (pass empty strings to skip auth).
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}, token, accountID string) error {
	url := c.baseURL + path

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("version", c.version)
	req.Header.Set("product", c.product)

	// Set auth headers if provided
	if token != "" && accountID != "" {
		c.setAuthHeader(req, token, accountID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &NetworkError{Err: err}
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle HTTP status codes
	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		// Success - decode response if result is provided
		if result != nil {
			if err := json.Unmarshal(respBody, result); err != nil {
				return fmt.Errorf("failed to decode response: %w", err)
			}
		}
		return nil

	case resp.StatusCode == http.StatusUnauthorized:
		return &AuthError{StatusCode: resp.StatusCode, Body: respBody}

	case resp.StatusCode == http.StatusTooManyRequests:
		return &RateLimitError{StatusCode: resp.StatusCode, Body: respBody}

	case resp.StatusCode >= 500:
		return &ServerError{StatusCode: resp.StatusCode, Body: respBody}

	default:
		return &HTTPError{StatusCode: resp.StatusCode, Body: respBody}
	}
}

// setAuthHeader sets the Authorization header and account-id for authenticated requests.
func (c *Client) setAuthHeader(req *http.Request, token, accountID string) {
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("account-id", accountID)
}
