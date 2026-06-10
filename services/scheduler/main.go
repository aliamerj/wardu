package main

import (
	"context"
	"net"
	"os/signal"
	"syscall"

	"github.com/aliamerj/wardu/services/scheduler/server"
	"github.com/aliamerj/wardu/shared/env"
	"github.com/aliamerj/wardu/shared/logger"
	r "github.com/aliamerj/wardu/shared/rabbitmq"
	"github.com/rs/zerolog"
	grpcServer "google.golang.org/grpc"
)

func main() {
	log := logger.Setup("scheduler")
	log.Info().Msg("starting scheduler service")

	grpcAddr := env.GetString("SCHEDULER_GRPC_PORT", ":8081")
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatal().Err(err).Str("addr", grpcAddr).Msg("failed to listen on scheduler port")
	}

	rabbitmq, err := r.New()
	if err != nil {
		_ = lis.Close()
		log.Fatal().Err(err).Msg("failed to start to rabbitMQ")
	}

	grpc := grpcServer.NewServer(
		grpcServer.ChainUnaryInterceptor(logger.UnaryServerInterceptor(log)),
	)
	server.NewGrpc(grpc, rabbitmq)

	done := make(chan struct{}, 1)

	log.Info().Str("addr", lis.Addr().String()).Msg("scheduler gRPC server listening")
	go gracefulShutdown(log, grpc, rabbitmq, done)

	if err := grpc.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("scheduler gRPC server stopped unexpectedly")
	}

	<-done
	log.Info().Msg("scheduler service exited cleanly")
}

func gracefulShutdown(log zerolog.Logger, grpc *grpcServer.Server, rabbitmq *r.RabbitMQ, done chan<- struct{}) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Info().Msg("shutdown signal received, stopping scheduler service")
	stop()

	rabbitmq.Close()
	log.Info().Msg("rabbitMQ connection close")
	grpc.GracefulStop()
	log.Info().Msg("scheduler gRPC shutdown complete")
	done <- struct{}{}
}
