package server

import (
	"context"

	"github.com/aliamerj/wardu/shared/database"
	"github.com/aliamerj/wardu/shared/env"
	"github.com/aliamerj/wardu/shared/k8s"
	pb "github.com/aliamerj/wardu/shared/proto/dispatcher"
	r "github.com/aliamerj/wardu/shared/rabbitmq"
	zlog "github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type gRPCServer struct {
	ctx      context.Context
	db       database.Service
	k8s      *k8s.Client
	rabbitmq *r.RabbitMQ
	sem      chan struct{}
	pb.UnimplementedDispatcherServiceServer
}

func NewGrpc(ctx context.Context, server *grpc.Server, rabbitmq *r.RabbitMQ) *gRPCServer {
	srv := &gRPCServer{
		db:       database.New(),
		k8s:      k8s.New(),
		rabbitmq: rabbitmq,
		ctx:      ctx,
		sem:      make(chan struct{}, env.GetInt("MAX_WORKERS", 20)),
	}
	pb.RegisterDispatcherServiceServer(server, srv)
	zlog.Info().Msg("registered dispatcher gRPC service")
	return srv
}

func (g *gRPCServer) StartJobConsumer() {
	err := g.rabbitmq.ConsumeJobs(
		g.ctx,
		func(jm r.JobMessage) error {
			g.sem <- struct{}{}
			go func() {
				defer func() {
					<-g.sem
				}()

				if err := g.executeJob(jm); err != nil {
					zlog.Error().
						Err(err).
						Str("job_id", jm.JobID).
						Msg("job execution failed")
				}
			}()

			return nil
		},
	)
	if err != nil {
		zlog.Fatal().
			Err(err).
			Msg("job consumer stopped")
	}
}

func (g *gRPCServer) RunJob(
	ctx context.Context,
	req *pb.RunJobRequest,
) (*pb.RunJobResponse, error) {
	err := g.rabbitmq.PublishJob(ctx, r.JobMessage{
		JobID:     req.GetJobId(),
		Priority:  req.GetPriority(),
		Attempt:   1,
		Overrides: buildOverridesOps(req),
	})
	if err != nil {
		zlog.Error().
			Err(err).
			Msg("job consumer stopped")
		return nil, err
	}

	return &pb.RunJobResponse{
		JobId: req.GetJobId(),
	}, nil
}

func buildOverridesOps(
	req *pb.RunJobRequest,
) *r.JobOverrides {
	var ops r.JobOverrides
	var hasOverrides bool

	if len(req.GetPayload()) > 0 {
		ops.Payload = req.GetPayload()
		hasOverrides = true
	}

	if len(req.GetEntrypoint()) > 0 {
		ops.Entrypoint = req.GetEntrypoint()
		hasOverrides = true
	}

	if req.IdleTimeoutSeconds != nil {
		v := req.GetIdleTimeoutSeconds()
		ops.IdleTimeoutSeconds = &v
		hasOverrides = true
	}

	if req.MaxAttempts != nil {
		v := req.GetMaxAttempts()
		ops.MaxAttempts = &v
		hasOverrides = true
	}

	if req.TimeoutSeconds != nil {
		v := req.GetTimeoutSeconds()
		ops.TimeoutSeconds = &v
		hasOverrides = true
	}

	if !hasOverrides {
		return nil
	}

	return &ops
}
