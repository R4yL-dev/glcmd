package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/R4yL-dev/glcmd/internal/daemon"
	"github.com/R4yL-dev/glcmd/internal/service"
)

// Server represents the HTTP API server
type Server struct {
	httpServer        *http.Server
	port              int
	glucoseService    service.GlucoseService
	sensorService     service.SensorService
	configService     service.ConfigService
	logger            *slog.Logger
	getHealthStatus   func() daemon.HealthStatus
	getDatabaseHealth func() bool
	startTime         time.Time
}

// NewServer creates a new API server instance
func NewServer(
	port int,
	glucoseService service.GlucoseService,
	sensorService service.SensorService,
	configService service.ConfigService,
	getHealthStatus func() daemon.HealthStatus,
	getDatabaseHealth func() bool,
	logger *slog.Logger,
) *Server {
	s := &Server{
		port:              port,
		glucoseService:    glucoseService,
		sensorService:     sensorService,
		configService:     configService,
		getHealthStatus:   getHealthStatus,
		getDatabaseHealth: getDatabaseHealth,
		startTime:         time.Now(),
		logger:            logger,
	}

	router := s.setupRouter()

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return s
}

// setupRouter configures the chi router with routes and middleware
func (s *Server) setupRouter() *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(s.corsMiddleware)      // CORS must be first for preflight requests
	r.Use(s.loggingMiddleware)
	r.Use(s.recoveryMiddleware)
	r.Use(s.timeoutMiddleware)

	// Monitoring endpoints at root (no versioning for operational endpoints)
	r.Get("/health", s.handleHealth)
	r.Get("/metrics", s.handleMetrics)

	// API v1 routes (versioned)
	r.Route("/v1", func(r chi.Router) {
		// Glucose routes
		r.Get("/glucose", s.handleGetGlucose)
		r.Get("/glucose/latest", s.handleGetLatestGlucose)
		r.Get("/glucose/stats", s.handleGetGlucoseStatistics)

		// Sensor routes
		r.Get("/sensor", s.handleGetSensor)
		r.Get("/sensor/latest", s.handleGetLatestSensor)
		r.Get("/sensor/stats", s.handleGetSensorStatistics)
	})

	return r
}

// Start starts the HTTP server in a goroutine
func (s *Server) Start() error {
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("API server error", "error", err)
		}
	}()
	return nil
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping API server")
	return s.httpServer.Shutdown(ctx)
}

// HTTPHandler returns the HTTP handler for testing purposes
func (s *Server) HTTPHandler() http.Handler {
	return s.httpServer.Handler
}
