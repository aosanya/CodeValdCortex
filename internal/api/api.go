package api

import (
	"context"
	"fmt"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/configuration"
	"github.com/aosanya/CodeValdCortex/internal/lifecycle"
	"github.com/aosanya/CodeValdCortex/internal/memory"
	"github.com/aosanya/CodeValdCortex/internal/templates"
)

// DefaultServerConfig returns a default server configuration
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Host:         "0.0.0.0",
		Port:         8080,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
		MaxBodySize:  10 * 1024 * 1024, // 10MB
		TLSEnabled:   false,
		TLSCertFile:  "",
		TLSKeyFile:   "",
		Environment:  "development",
	}
}

// NewDefaultServices creates default service instances for development
func NewDefaultServices() (*Services, error) {
	// Configuration service - using nil implementations for now
	configService := configuration.NewService(nil, nil, nil, nil)

	// Template engine - using nil implementations for now
	templateEngine := templates.NewEngine(nil, nil)

	// Memory service (using in-memory implementation for now)
	memoryService := memory.NewService(nil) // nil for in-memory

	// Lifecycle manager (will need repository implementation)
	lifecycleManager := lifecycle.NewManager(nil) // nil for now

	return &Services{
		ConfigService:    configService,
		TemplateEngine:   templateEngine,
		LifecycleManager: lifecycleManager,
		MemoryService:    memoryService,
	}, nil
}

// StartServer is a convenience function to start the API server with default settings
func StartServer(ctx context.Context, config *ServerConfig) error {
	if config == nil {
		config = DefaultServerConfig()
	}

	services, err := NewDefaultServices()
	if err != nil {
		return fmt.Errorf("failed to initialize services: %w", err)
	}

	server := NewServer(config, services)

	// Start server in goroutine
	go func() {
		if err := server.Start(); err != nil {
			fmt.Printf("Server failed to start: %v\n", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return server.Stop(shutdownCtx)
}
