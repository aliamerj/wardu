package client

import (
	"context"

	"github.com/aliamerj/wardu/shared/env"
	"github.com/aliamerj/wardu/shared/logger"
	pb "github.com/aliamerj/wardu/shared/proto/dispatcher"
	zlog "github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Dispatcher struct {
	client pb.DispatcherServiceClient
	conn   *grpc.ClientConn
}

func NewDispatcher() (*Dispatcher, error) {
	addr := env.GetString("DISPATCHER_GRPC_ADDR", "localhost:8082")

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(logger.UnaryClientInterceptor(zlog.Logger)),
	)
	if err != nil {
		return nil, err
	}

	client := pb.NewDispatcherServiceClient(conn)
	zlog.Info().Str("addr", addr).Msg("dispatcher gRPC client connected")

	return &Dispatcher{conn: conn, client: client}, nil
}

func (d *Dispatcher) Heartbeat(
	ctx context.Context,
) (pb.DispatcherService_HeartbeatClient, error) {
	return d.client.Heartbeat(ctx)
}

func (s *Dispatcher) Close() error {
	if s.conn != nil {
		if err := s.conn.Close(); err != nil {
			return err
		}
		zlog.Info().Msg("scheduler gRPC client closed")
	}
	return nil
}
