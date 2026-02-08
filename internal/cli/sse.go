package cli

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strings"
)

// SSEEvent represents a parsed SSE event
type SSEEvent struct {
	Type string
	Data []byte
}

// Stream connects to the SSE endpoint and returns channels for events and errors.
// types specifies which event types to receive (empty = all).
// The returned channels are closed when the context is cancelled or an error occurs.
func (c *Client) Stream(ctx context.Context, types []string) (<-chan SSEEvent, <-chan error) {
	events := make(chan SSEEvent)
	errors := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errors)

		err := c.streamEvents(ctx, types, events)
		if err != nil && ctx.Err() == nil {
			errors <- err
		}
	}()

	return events, errors
}

// streamEvents handles the SSE connection and event parsing
func (c *Client) streamEvents(ctx context.Context, types []string, events chan<- SSEEvent) error {
	// Build URL with type filter
	path := "/v1/stream"
	if len(types) > 0 {
		path = fmt.Sprintf("/v1/stream?types=%s", strings.Join(types, ","))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "text/event-stream")

	// Use a client without timeout for streaming
	streamClient := &http.Client{} // No timeout for SSE
	resp, err := streamClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot connect to glcore at %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SSE endpoint returned status %d", resp.StatusCode)
	}

	// Read SSE events
	reader := bufio.NewReader(resp.Body)
	var currentEvent SSEEvent

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("connection closed: %w", err)
		}

		line = strings.TrimSuffix(line, "\n")
		line = strings.TrimSuffix(line, "\r")

		if line == "" {
			// Empty line = event boundary
			if currentEvent.Type != "" {
				select {
				case events <- currentEvent:
				case <-ctx.Done():
					return nil
				}
			}
			currentEvent = SSEEvent{}
			continue
		}

		if strings.HasPrefix(line, "event:") {
			currentEvent.Type = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			data := strings.TrimPrefix(line, "data:")
			data = strings.TrimSpace(data)
			currentEvent.Data = []byte(data)
		}
		// Ignore comments (lines starting with :) and other fields
	}
}
