package clients

import (
	"context"

	"github.com/aliamerj/wardu/shared/env"
	pb "github.com/aliamerj/wardu/shared/proto/scheduler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type schedulerClient struct {
	client pb.SchedulerServiceClient
	conn   *grpc.ClientConn
}

func newScheduler() (*schedulerClient, error) {
	addr := env.GetString("SCHEDULER_GRPC_ADDR", "localhost:8081")

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := pb.NewSchedulerServiceClient(conn)

	return &schedulerClient{
		conn:   conn,
		client: client,
	}, nil
}

func (s *schedulerClient) CreateJob(ctx context.Context, job *pb.CreateJobRequest) (*pb.CreateJobResponse, error) {
	return s.client.CreateJob(ctx, job)
}

func (s *schedulerClient) close() error {
	if s.conn != nil {
		if err := s.conn.Close(); err != nil {
			return err
		}
	}
	return nil
}
