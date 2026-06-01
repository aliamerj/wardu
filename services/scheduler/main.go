package main

import (
	"context"
	"log"
	"net"
	"os/signal"
	"syscall"

	"github.com/aliamerj/wardu/services/scheduler/server"
	"github.com/aliamerj/wardu/shared/env"
	grpcServer "google.golang.org/grpc"
)

func main() {
	grpcAddr := env.GetString("SCHEDULER_GRPC_PORT", ":8081")

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpc := grpcServer.NewServer()
	server.NewGrpc(grpc)

	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	log.Printf("Starting gRPC server Scheduler on port : %s", lis.Addr().String())

	// Run graceful shutdown in a separate goroutine
	go gracefulShutdown(grpc, done)

	if err := grpc.Serve(lis); err != nil {
		log.Fatalf("gRPC server Scheduler server error: %s", err)
	}

	// Wait for the graceful shutdown to complete
	<-done

	log.Println("Graceful shutdown gRPC server Scheduler complete.")
}

func gracefulShutdown(grpc *grpcServer.Server, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Println("Shutting down gRPC server Scheduler gracefully, press Ctrl+C again to force")
	stop() // Allow Ctrl+C to force shutdown

	grpc.GracefulStop()

	log.Println("Scheduler service exiting")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}
