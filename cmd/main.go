package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hojamuhammet/go-grpc-otp-rabbitmq/internal/pkg/config"
	"github.com/hojamuhammet/go-grpc-otp-rabbitmq/internal/pkg/server"
	"github.com/lib/pq"
)

func main() {
    // Load configuration from your configuration file or environment variables
    cfg := config.LoadConfig()

	dbURL := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
		
    // Initialize a PostgreSQL database connection pool
    db, err := pq.Open(dbURL)
    if err != nil {
        log.Fatalf("Failed to connect to the database: %v", err)
    }
    defer db.Close()

    // Create a new gRPC server instance
    grpcServer := server.NewServer()

    // Start the gRPC server in a separate goroutine
    go func() {
        if err := grpcServer.Start(cfg.GRPCPort); err != nil {
            log.Fatalf("Failed to start gRPC server: %v", err)
        }
    }()

    // Handle graceful shutdown on SIGINT and SIGTERM signals
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

    select {
    case <-sigCh:
        log.Println("Received termination signal. Shutting down...")
        grpcServer.Stop() // Gracefully stop the gRPC server
        grpcServer.Wait() // Wait for the server to finish gracefully
    }
}
