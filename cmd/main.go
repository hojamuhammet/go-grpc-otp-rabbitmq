package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hojamuhammet/go-grpc-otp-rabbitmq/internal/config"
	"github.com/hojamuhammet/go-grpc-otp-rabbitmq/internal/database"
	"github.com/hojamuhammet/go-grpc-otp-rabbitmq/internal/rabbitmq"
	server "github.com/hojamuhammet/go-grpc-otp-rabbitmq/internal/server"
	"github.com/joho/godotenv"
)

func main() {
    if err := godotenv.Load(); err != nil {
        log.Fatalf("Error loading the env variables: %v", err)
    }

    cfg := config.LoadConfig()
    
	dbURL := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
		
    // Initialize a PostgreSQL database connection pool
    db, err := database.NewDatabase(dbURL)
    if err != nil {
        log.Fatalf("Failed to connect to the database: %v", err)
    }
    defer db.Close()

    // Initialize RabbitMQ service
    rabbitMQService, err := rabbitmq.InitRabbitMQConnection(cfg.RabbitMQ_URL)
    if err != nil {
        log.Fatalf("Failed to initialize RabbitMQ service: %v", err)
    }
    defer rabbitMQService.Close()

    // Create a new gRPC server instance
    grpcServer := server.NewServer(&cfg, db, rabbitMQService)

    // Start the gRPC server in a separate goroutine
    go func() {
        if err := grpcServer.Start(context.Background(), &cfg); err != nil {
            log.Fatalf("Failed to start gRPC server: %v", err)
        }
    }()

    // Handle graceful shutdown on SIGINT and SIGTERM signals
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

    <-sigCh
    log.Println("Received termination signal. Shutting down...")
    grpcServer.Stop() // Gracefully stop the gRPC server
    grpcServer.Wait() // Wait for the server to finish gracefully
}
