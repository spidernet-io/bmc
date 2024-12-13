package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()
	log := logger.Sugar()

	log.Info("Starting BMC agent")

	// Get cluster name from environment variable
	clusterName := os.Getenv("ClusterName")
	if clusterName == "" {
		log.Error("ClusterName environment variable not set")
		os.Exit(1)
	}
	log.Infof("Running agent for cluster: %s", clusterName)

	// Setup HTTP server for health checks
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Health check request received")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "healthy")
	})

	// Readiness check endpoint
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Readiness check request received")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ready")
	})

	server := &http.Server{
		Addr:    ":8000",
		Handler: mux,
	}

	// Start HTTP server in a goroutine
	go func() {
		log.Info("Starting health check server on :8000")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Errorf("Health check server failed: %v", err)
			os.Exit(1)
		}
	}()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// Main loop - sleep and log periodically
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Debug("Agent still running...")
		case sig := <-sigChan:
			log.Infof("Received signal %v, shutting down...", sig)
			return
		}
	}
}
