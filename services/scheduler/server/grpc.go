package server

import (
	"context"

	"github.com/aliamerj/wardu/services/scheduler/handlers"
	"github.com/aliamerj/wardu/shared/database"
	"github.com/aliamerj/wardu/shared/k8s"
	pb "github.com/aliamerj/wardu/shared/proto/scheduler"
	"google.golang.org/grpc"
)

type gRPCServer struct {
	h *handlers.Handler
	pb.UnimplementedSchedulerServiceServer
}

func NewGrpc(server *grpc.Server) *gRPCServer {
	srv := &gRPCServer{
		h: handlers.New(database.New(), k8s.New()),
	}
	pb.RegisterSchedulerServiceServer(server, srv)
	return srv
}

func (g *gRPCServer) CreateJob(ctx context.Context, req *pb.CreateJobRequest) (*pb.CreateJobResponse, error) {
	jobID, err := g.h.CreateJob(ctx, req)
	if err != nil {
		return nil, err
	}
	return &pb.CreateJobResponse{
		JobId: jobID,
	}, nil
}
