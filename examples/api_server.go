package examples
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/aosanya/CodeValdCortex/internal/api"
)

func main() {
	// Command line flags
	var (
		host    = flag.String("host", "0.0.0.0", "Host to bind the API server to")
		port    = flag.Int("port", 8080, "Port to bind the API server to")
		env     = flag.String("env", "development", "Environment (development, production)")
		debug   = flag.Bool("debug", false, "Enable debug logging")
	)
	flag.Parse()

	// Configure logging
	if *debug || *env == "development" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	log.WithFields(log.Fields{
		"host": *host,
		"port": *port,
		"env":  *env,
	}).Info("Starting CodeValdCortex API Server")

	// Create server configuration
	config := &api.ServerConfig{
		Host:         *host,
		Port:         *port,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
		MaxBodySize:  10 * 1024 * 1024, // 10MB
		TLSEnabled:   false,
		Environment:  *env,
	}

	// Create context that listens for interrupt signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown gracefully
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		log.Info("Received shutdown signal")
		cancel()
	}()

	// Start the API server
	if err := api.StartServer(ctx, config); err != nil {
		log.WithError(err).Fatal("Failed to start API server")
	}

	log.Info("API Server stopped")
}