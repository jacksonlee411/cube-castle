package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cube-castle/cmd/organization-command-server/internal/infrastructure/config"
	"github.com/cube-castle/cmd/organization-command-server/internal/infrastructure/container"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize dependency container
	appContainer, err := container.NewContainer(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}
	defer appContainer.Close()

	// Get HTTP server
	server := appContainer.GetHTTPServer()

	// Setup graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	// Start server
	log.Printf("🚀 Organization Command Service v%s starting on port %s", 
		cfg.App.Version, server.Addr)
	log.Printf("Environment: %s", cfg.App.Environment)

	if err := server.ListenAndServe(); err != nil {
		log.Printf("Server stopped: %v", err)
	}

	log.Println("Server shutdown complete")
}