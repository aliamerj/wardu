package clients

import (
	"context"

	"github.com/aliamerj/wardu/shared/env"
	"github.com/aliamerj/wardu/shared/logger"
	pb "github.com/aliamerj/wardu/shared/proto/scheduler"
	zlog "github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type schedulerClient struct {
	client pb.SchedulerServiceClient
	conn   *grpc.ClientConn
}

func newScheduler() (*schedulerClient, error) {
	addr := env.GetString("SCHEDULER_GRPC_ADDR", "localhost:8081")

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(logger.UnaryClientInterceptor(zlog.Logger)),
	)
	if err != nil {
		return nil, err
	}

	client := pb.NewSchedulerServiceClient(conn)
	zlog.Info().Str("addr", addr).Msg("scheduler gRPC client connected")

	return &schedulerClient{conn: conn, client: client}, nil
}

func (s *schedulerClient) CreateJob(ctx context.Context, job *pb.CreateJobRequest) (*pb.CreateJobResponse, error) {
	return s.client.CreateJob(ctx, job)
}

func (s *schedulerClient) close() error {
	if s.conn != nil {
		if err := s.conn.Close(); err != nil {
			return err
		}
		zlog.Info().Msg("scheduler gRPC client closed")
	}
	return nil
}
