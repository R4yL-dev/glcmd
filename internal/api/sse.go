package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/R4yL-dev/glcmd/internal/events"
)

// handleSSEStream handles GET /v1/stream
// Query params: types=glucose,sensor (optional, default = all)
func (s *Server) handleSSEStream(w http.ResponseWriter, r *http.Request) {
	// Check if SSE is enabled (broker is set)
	if s.eventBroker == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "SSE streaming not available")
		return
	}

	// Check for streaming support
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "SSE not supported")
		return
	}

	// Disable write timeout for SSE (long-lived connection)
	rc := http.NewResponseController(w)
	if err := rc.SetWriteDeadline(time.Time{}); err != nil {
		s.logger.Warn("failed to disable write deadline for SSE", "error", err)
	}

	// Parse type filter from query params
	types := parseEventTypes(r.URL.Query().Get("types"))

	// Generate client ID
	clientID := uuid.New().String()
	start := time.Now()

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering

	// Log connection
	s.logger.Info("SSE client connected",
		"clientID", clientID,
		"path", r.URL.Path,
		"types", types,
		"subscribers", s.eventBroker.SubscriberCount()+1,
	)

	// Subscribe to events
	eventCh := s.eventBroker.Subscribe(clientID, types)
	defer func() {
		s.eventBroker.Unsubscribe(clientID)
		s.logger.Info("SSE client disconnected",
			"clientID", clientID,
			"duration", time.Since(start),
			"subscribers", s.eventBroker.SubscriberCount(),
		)
	}()

	// Flush headers immediately
	flusher.Flush()

	// Stream events
	for {
		select {
		case event, ok := <-eventCh:
			if !ok {
				// Channel closed, broker stopped
				return
			}
			if err := writeSSEEvent(w, flusher, event); err != nil {
				// Client disconnected
				return
			}
		case <-r.Context().Done():
			// Client disconnected
			return
		}
	}
}

// parseEventTypes parses comma-separated event types from query string
func parseEventTypes(typesParam string) []events.EventType {
	if typesParam == "" {
		return nil // Empty = all types
	}

	parts := strings.Split(typesParam, ",")
	types := make([]events.EventType, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		switch p {
		case "glucose":
			types = append(types, events.EventTypeGlucose)
		case "sensor":
			types = append(types, events.EventTypeSensor)
		case "keepalive":
			types = append(types, events.EventTypeKeepalive)
		}
	}

	return types
}

// writeSSEEvent writes a single SSE event to the response
func writeSSEEvent(w http.ResponseWriter, flusher http.Flusher, event events.Event) error {
	var data []byte
	var err error

	if event.Data != nil {
		data, err = json.Marshal(event.Data)
		if err != nil {
			data = []byte("{}")
		}
	} else {
		data = []byte("{}")
	}

	// Write event in SSE format:
	// event: <type>
	// data: <json>
	// (blank line)
	_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, data)
	if err != nil {
		return err
	}

	flusher.Flush()
	return nil
}
