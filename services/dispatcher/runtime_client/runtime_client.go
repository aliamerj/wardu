package client

import (
	"context"

	"github.com/aliamerj/wardu/shared/logger"
	pb "github.com/aliamerj/wardu/shared/proto/runtime"
	zlog "github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Runtime struct {
	addr   string
	client pb.RuntimeServiceClient
	conn   *grpc.ClientConn
}

func NewRuntime(addr string) (*Runtime, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(logger.UnaryClientInterceptor(zlog.Logger)),
	)
	if err != nil {
		return nil, err
	}

	client := pb.NewRuntimeServiceClient(conn)
	zlog.Info().Str("addr", addr).Msg("runtime gRPC client connected")

	return &Runtime{addr: addr, conn: conn, client: client}, nil
}

func (r *Runtime) Run(ctx context.Context, req *pb.RunRequest) (*pb.RunResponse, error) {
	return r.client.Run(ctx, req)
}

func (s *Runtime) Close() {
	if s.conn != nil {
		if err := s.conn.Close(); err != nil {
			zlog.Error().Err(err).Str("addr", s.addr).Msg("Failed to close runtime")
			return
		}
		zlog.Info().Msg("runtime gRPC client closed")
	}
	return
}
