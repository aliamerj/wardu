package main

import (
	"context"
	"net"
	"os/signal"
	"syscall"

	"github.com/aliamerj/wardu/services/dispatcher/server"
	"github.com/aliamerj/wardu/shared/env"
	"github.com/aliamerj/wardu/shared/logger"
	r "github.com/aliamerj/wardu/shared/rabbitmq"
	"github.com/rs/zerolog"
	grpcServer "google.golang.org/grpc"
)

func main() {
	log := logger.Setup("dispatcher")
	log.Info().Msg("starting dispatcher service")

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	grpcAddr := env.GetString("DISPATCHER_GRPC_PORT", ":8082")

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("addr", grpcAddr).
			Msg("failed to listen on dispatcher port")
	}

	rabbitmq, err := r.New()
	if err != nil {
		if err = lis.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close gRPC connection")
		}
		log.Fatal().
			Err(err).
			Msg("failed to connect to rabbitmq")
	}

	grpc := grpcServer.NewServer(
		grpcServer.ChainUnaryInterceptor(
			logger.UnaryServerInterceptor(log),
		),
	)

	srv := server.NewGrpc(ctx, grpc, rabbitmq)

	go srv.StartJobConsumer()

	go gracefulShutdown(
		ctx,
		log,
		grpc,
		rabbitmq,
	)

	log.Info().
		Str("addr", lis.Addr().String()).
		Msg("dispatcher gRPC server listening")

	if err := grpc.Serve(lis); err != nil {
		log.Fatal().
			Err(err).
			Msg("dispatcher gRPC server stopped unexpectedly")
	}
}

func gracefulShutdown(
	ctx context.Context,
	log zerolog.Logger,
	grpc *grpcServer.Server,
	rabbitmq *r.RabbitMQ,
) {
	<-ctx.Done()

	log.Info().
		Msg("shutdown signal received")

	rabbitmq.Close()

	log.Info().
		Msg("rabbitmq connection closed")

	log.Info().Msg("stopping grpc server")

	grpc.GracefulStop()

	log.Info().Msg("dispatcher gRPC shutdown complete")
}
