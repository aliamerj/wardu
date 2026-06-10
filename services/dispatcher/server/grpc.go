package server

import (
	"context"

	"github.com/aliamerj/wardu/services/dispatcher/handlers"
	"github.com/aliamerj/wardu/shared/database"
	"github.com/aliamerj/wardu/shared/k8s"
	pb "github.com/aliamerj/wardu/shared/proto/dispatcher"
	r "github.com/aliamerj/wardu/shared/rabbitmq"
	zlog "github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type gRPCServer struct {
	db       database.Service
	k8s      *k8s.Client
	rabbitmq *r.RabbitMQ
	pb.UnimplementedDispatcherServiceServer
}

func NewGrpc(server *grpc.Server, rabbitmq *r.RabbitMQ) *gRPCServer {
	srv := &gRPCServer{
		db:       database.New(),
		k8s:      k8s.New(),
		rabbitmq: rabbitmq,
	}
	pb.RegisterDispatcherServiceServer(server, srv)
	zlog.Info().Msg("registered dispatcher gRPC service")
	return srv
}

func (g *gRPCServer) StartJobConsumer(ctx context.Context) {
	err := g.rabbitmq.ConsumeJobs(ctx, func(jm r.JobMessage) error {
		return handlers.ExecuteJob(ctx, g.db, g.k8s, jm)
	})
	if err != nil {
		zlog.Fatal().
			Err(err).
			Msg("job consumer stopped")
	}
}

func (g *gRPCServer) RunJob(ctx context.Context, req *pb.RunJobRequest) (*pb.RunJobResponse, error) {
	return &pb.RunJobResponse{
		JobId: "",
	}, nil
}
