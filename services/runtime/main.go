package main

import (
	"context"
	"net"
	"os/signal"
	"syscall"

	client "github.com/aliamerj/wardu/services/runtime/dispatcher_client"
	"github.com/aliamerj/wardu/services/runtime/server"
	"github.com/aliamerj/wardu/shared/env"
	"github.com/aliamerj/wardu/shared/logger"
	"github.com/rs/zerolog"
	grpcServer "google.golang.org/grpc"
)

func main() {
	log := logger.Setup("runtime")
	log.Info().Msg("starting runtime service")

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	grpcAddr := env.GetString("RUNTIME_GRPC_PORT", ":8083")

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("addr", grpcAddr).
			Msg("failed to listen on runtime port")
	}

	dispatcher, err := client.NewDispatcher()
	if err != nil {
		_ = lis.Close()

		log.Fatal().
			Err(err).
			Msg("failed to connect to dispatcher")
	}

	grpc := grpcServer.NewServer(
		grpcServer.ChainUnaryInterceptor(
			logger.UnaryServerInterceptor(log),
		),
	)

	server.NewGrpc(ctx, grpc, dispatcher)

	go gracefulShutdown(
		ctx,
		log,
		grpc,
		dispatcher,
	)

	log.Info().
		Str("addr", lis.Addr().String()).
		Msg("runtime gRPC server listening")

	if err := grpc.Serve(lis); err != nil {
		log.Fatal().
			Err(err).
			Msg("runtime gRPC server stopped unexpectedly")
	}
}

func gracefulShutdown(
	ctx context.Context,
	log zerolog.Logger,
	grpc *grpcServer.Server,
	dispatcher *client.Dispatcher,
) {
	<-ctx.Done()

	log.Info().
		Msg("shutdown signal received")

	if err := dispatcher.Close(); err != nil {
		log.Error().Err(err).Msg("failed to close dispatcher client connections")
	}

	grpc.GracefulStop()

	log.Info().
		Msg("runtime gRPC shutdown complete")
}
