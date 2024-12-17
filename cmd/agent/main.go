package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"github.com/spidernet-io/bmc/pkg/log"

)

func main() {
	// Get log level from environment variable
	logLevel := os.Getenv("LOG_LEVEL")
	log.InitStdoutLogger(logLevel)


	log.Logger.Infof("Starting BMC agent")

	// Get cluster name from environment variable
	clusterName := os.Getenv("CLUSTER_NAME")
	if clusterName == "" {
		log.Logger.Errorf("CLUSTER_NAME environment variable not set")
		os.Exit(1)
	}
	log.Logger.Infof("Running agent for cluster: %s", clusterName)

	// Setup HTTP server for health checks
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

	server := &http.Server{
		Addr:    ":8000",
		Handler: mux,
	}

	// Start HTTP server in a goroutine
	go func() {
		log.Logger.Info("Starting health check server on :8000")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Logger.Errorf("Health check server failed: %v", err)
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
			log.Logger.Debugf("Agent still running...")
		case sig := <-sigChan:
			log.Logger.Infof("Received signal %v, shutting down...", sig)
			return
		}
	}
}
