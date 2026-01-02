package healthcheck

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"time"
)

// HealthStatus is an interface for health status data.
// Use daemon.HealthStatus which implements this interface.
type HealthStatus interface{}

// Server provides HTTP endpoints for health checks and metrics.
type Server struct {
	httpServer      *http.Server
	port            int
	startTime       time.Time
	getHealthStatus func() interface{} // Callback to get current health status from daemon
}

// NewServer creates a new healthcheck server.
//
// Parameters:
//   - port: The port to listen on (e.g., 8080)
//   - getHealthStatus: Callback function that returns current health status
func NewServer(port int, getHealthStatus func() interface{}) *Server {
	s := &Server{
		port:            port,
		startTime:       time.Now(),
		getHealthStatus: getHealthStatus,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/metrics", s.handleMetrics)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return s
}

// Start starts the HTTP server in a goroutine.
func (s *Server) Start() error {
	go func() {
		slog.Info("healthcheck server starting", "port", s.port)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("healthcheck server failed", "error", err)
		}
	}()
	return nil
}

// Stop gracefully shuts down the HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	slog.Info("shutting down healthcheck server")
	return s.httpServer.Shutdown(ctx)
}

// handleHealth handles GET /health requests.
//
// Returns:
//   - 200 OK if status is "healthy"
//   - 503 Service Unavailable if status is "degraded" or "unhealthy"
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	statusData := s.getHealthStatus()

	w.Header().Set("Content-Type", "application/json")

	// Encode and determine status code based on the json "status" field
	// This works because statusData implements json.Marshaler via struct tags
	data, err := json.Marshal(statusData)
	if err != nil {
		slog.Error("failed to marshal health status", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Parse to check status field
	var statusMap map[string]interface{}
	if err := json.Unmarshal(data, &statusMap); err == nil {
		if status, ok := statusMap["status"].(string); ok {
			if status == "degraded" || status == "unhealthy" {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write(data)
				return
			}
		}
	}

	// Default to healthy (200 OK)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// handleMetrics handles GET /metrics requests.
//
// Returns basic daemon metrics in JSON format.
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := map[string]interface{}{
		"uptime":           time.Since(s.startTime).String(),
		"goroutines":       runtime.NumGoroutine(),
		"memory": map[string]interface{}{
			"allocMB":      m.Alloc / 1024 / 1024,
			"totalAllocMB": m.TotalAlloc / 1024 / 1024,
			"sysMB":        m.Sys / 1024 / 1024,
			"numGC":        m.NumGC,
		},
		"runtime": map[string]interface{}{
			"version": runtime.Version(),
			"os":      runtime.GOOS,
			"arch":    runtime.GOARCH,
		},
		"process": map[string]interface{}{
			"pid": os.Getpid(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		slog.Error("failed to encode metrics", "error", err)
	}
}
