package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/spidernet-io/bmc/pkg/log"
)

// Server represents the HTTP server
type Server struct {
	httpServer *http.Server
}

// NewServer creates a new HTTP server
func NewServer(port int32) *Server {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		log.Logger.Debugf("Health check request received")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "healthy")
	})

	// Readiness check endpoint
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		log.Logger.Debugf("Readiness check request received")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ready")
	})

	return &Server{
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		},
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	log.Logger.Infof("Starting health check server on %s", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("health check server failed: %v", err)
	}
	return nil
}

// Shutdown gracefully shuts down the HTTP server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
