// Package httpserver implements helper functions for creating http servers.
package httpserver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	DefayltPort              = 8000 //nolint:revive
	DefaultLivenessProbePath = "/alive"
	DefaultReadHeaderTimeout = 5 * time.Second
)

// ServerConfig holds the configuration for the http server
type ServerConfig struct {
	// Logger is the logger used by the server
	Logger *slog.Logger
	// Port is the port the server listens on. Default to DefaultPort
	Port int
	// EnableMetrics enables the prometheus metrics handler at the /metrics route
	EnableMetrics bool
	// LivenessProbe enables the liveness probe handler
	LivenessProbe bool
	// LivenessProbePath is the path for the liveness probe handler. Default is DefaultLivenessProbePath
	LivenessProbePath string
	// ReadHeaderTimeout is the maximum duration before timing out read of the request headers.
	// Defaults to DefaultReadHeaderTimeout
	ReadHeaderTimeout time.Duration
}

// Server is a http server that implements common requirements such as liveness probe, exposing metrics and
// graceful shutdown
type Server struct {
	srv               *http.ServeMux
	log               *slog.Logger
	port              int
	readHeaderTimeout time.Duration
	shutdownTimeout   time.Duration
}

// livenessHandler is a simple handler that returns a 200 status code.
func livenessHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// NewServer creates a new http server with the given configuration.
func NewServer(config ServerConfig) *Server {
	srv := http.NewServeMux()
	if config.EnableMetrics {
		srv.Handle("/metrics", promhttp.Handler())
	}

	if config.LivenessProbe {
		livenessProbePath := config.LivenessProbePath
		if livenessProbePath == "" {
			livenessProbePath = DefaultLivenessProbePath
		}
		srv.HandleFunc(livenessProbePath, livenessHandler)
	}

	readHeaderTimeout := config.ReadHeaderTimeout
	if readHeaderTimeout == 0 {
		readHeaderTimeout = DefaultReadHeaderTimeout
	}

	log := config.Logger
	if log == nil {
		log = slog.New(slog.NewJSONHandler(io.Discard, nil))
	}

	return &Server{
		log:               log,
		port:              config.Port,
		srv:               srv,
		readHeaderTimeout: readHeaderTimeout,
		shutdownTimeout:   5 * time.Second,
	}
}

// Handle registers the handler for the given pattern.
func (s *Server) Handle(pattern string, handler http.Handler) {
	s.srv.Handle(pattern, handler)
}

// Start starts the http server and listens for incoming requests. It also listens for os signals to gracefully
// shutdown the server.
func (s *Server) Start(ctx context.Context) error {
	serverErrors := make(chan error, 1)

	srv := http.Server{
		Addr:              fmt.Sprintf(":%d", s.port),
		Handler:           s.srv,
		ReadHeaderTimeout: s.readHeaderTimeout,
	}

	go func() {
		s.log.Info("starting server", "address", srv.Addr)
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		s.log.Debug("shutdown started", "signal", sig)

		ctx, cancel := context.WithTimeout(ctx, s.shutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			s.log.Error("graceful shutdown failed", "error", err)
			if err := srv.Close(); err != nil {
				return fmt.Errorf("could not stop server: %w", err)
			}
		}

		s.log.Debug("shutdown completed")
	}

	return nil
}
