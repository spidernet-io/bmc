package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spidernet-io/bmc/pkg/agent/config"
	"github.com/spidernet-io/bmc/pkg/agent/server"
	"github.com/spidernet-io/bmc/pkg/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// Parse command line flags
	healthPort := flag.Int("health-port", 8080, "Port for health check server")
	flag.Parse()

	// Initialize logger
	logLevel := os.Getenv("LOG_LEVEL")
	log.InitStdoutLogger(logLevel)

	log.Logger.Info("Starting BMC agent")

	// Initialize Kubernetes client
	k8sClient, err := initClients()
	if err != nil {
		log.Logger.Errorf("Failed to initialize clients: %v", err)
		os.Exit(1)
	}

	// Load agent configuration
	agentConfig, err := config.LoadAgentConfig(k8sClient)
	if err != nil {
		log.Logger.Errorf("Failed to load agent configuration: %v", err)
		os.Exit(1)
	}

	// Validate endpoint configuration
	if err := agentConfig.ValidateEndpointConfig(k8sClient); err != nil {
		log.Logger.Errorf("Invalid endpoint configuration: %v", err)
		os.Exit(1)
	}

	// Validate feature configuration
	if err := agentConfig.ValidateFeatureConfig(); err != nil {
		log.Logger.Errorf("Invalid feature configuration: %v", err)
		os.Exit(1)
	}

	log.Logger.Info("Agent configuration loaded and validated successfully")

	// Create and start HTTP server for health checks
	srv := server.NewServer(int32(*healthPort))
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
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
			log.Logger.Debug("Agent still running...")
		case sig := <-sigChan:
			log.Logger.Infof("Received signal %v, shutting down...", sig)
			// Graceful shutdown of HTTP server
			if err := srv.Shutdown(context.Background()); err != nil {
				log.Logger.Errorf("Error shutting down HTTP server: %v", err)
			}
			return
		}
	}
}

// initClients initializes Kubernetes client
func initClients() (*kubernetes.Clientset, error) {
	// Get kubernetes config
	kubeconfig := os.Getenv("KUBECONFIG")
	var config *rest.Config
	var err error

	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, err
	}

	// Create kubernetes client
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return k8sClient, nil
}
