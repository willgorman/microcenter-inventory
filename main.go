package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/willgorman/microcenter-inventory/config"
	"github.com/willgorman/microcenter-inventory/inventory"
	"github.com/willgorman/microcenter-inventory/prometheus"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Parse command-line flags
	configFile := flag.String("config", "config.yaml", "Path to configuration file")
	metricsAddr := flag.String("metrics-addr", ":9090", "Address to expose metrics on")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Prometheus metrics
	prometheus.InitMetrics()

	// Create the inventory checker
	checker, err := inventory.NewChecker(cfg)
	if err != nil {
		log.Fatalf("Failed to create inventory checker: %v", err)
	}
	defer checker.Close()

	// Setup metrics HTTP server
	http.Handle("/metrics", promhttp.Handler())
	server := &http.Server{
		Addr: *metricsAddr,
	}

	// Setup graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Metrics server listening on %s", *metricsAddr)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Start the checker in a separate goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go checker.Start(ctx)

	// Wait for shutdown signal
	<-stop
	log.Println("Shutting down...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
}
