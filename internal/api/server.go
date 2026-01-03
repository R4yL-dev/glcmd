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
	httpServer      *http.Server
	port            int
	glucoseService  service.GlucoseService
	sensorService   service.SensorService
	configService   service.ConfigService
	logger          *slog.Logger
	getHealthStatus func() daemon.HealthStatus
	startTime       time.Time
}

// NewServer creates a new API server instance
func NewServer(
	port int,
	glucoseService service.GlucoseService,
	sensorService service.SensorService,
	configService service.ConfigService,
	getHealthStatus func() daemon.HealthStatus,
	logger *slog.Logger,
) *Server {
	s := &Server{
		port:            port,
		glucoseService:  glucoseService,
		sensorService:   sensorService,
		configService:   configService,
		getHealthStatus: getHealthStatus,
		startTime:       time.Now(),
		logger:          logger,
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
	r.Use(s.loggingMiddleware)
	r.Use(s.recoveryMiddleware)
	r.Use(s.timeoutMiddleware)

	// Health and metrics routes
	r.Get("/health", s.handleHealth)
	r.Get("/metrics", s.handleMetrics)

	// Measurement routes
	r.Get("/measurements", s.handleGetMeasurements)
	r.Get("/measurements/latest", s.handleGetLatestMeasurement)
	r.Get("/measurements/stats", s.handleGetStatistics)

	// Sensor routes
	r.Get("/sensors", s.handleGetSensors)

	return r
}

// Start starts the HTTP server in a goroutine
func (s *Server) Start() error {
	go func() {
		s.logger.Info("starting API server", "port", s.port)
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
