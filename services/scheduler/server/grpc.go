package server

import (
	"context"

	pb "github.com/aliamerj/wardu/shared/proto/scheduler"
	"google.golang.org/grpc"
)

type gRPCServer struct {
	pb.UnimplementedSchedulerServiceServer
}

func NewGrpc(server *grpc.Server) *gRPCServer {
	srv := &gRPCServer{}
	pb.RegisterSchedulerServiceServer(server, srv)
	return srv
}

func (h *gRPCServer) CreateJob(ctx context.Context, req *pb.CreateJobRequest) (*pb.CreateJobResponse, error) {
	return &pb.CreateJobResponse{
		JobId: req.JobId + " from grpc",
	}, nil
}
