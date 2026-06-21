package server

import (
	"context"
	"sync"
	"time"

	client "github.com/aliamerj/wardu/services/runtime/dispatcher_client"
	"github.com/aliamerj/wardu/shared/ids"
	dispatcherpb "github.com/aliamerj/wardu/shared/proto/dispatcher"
	pb "github.com/aliamerj/wardu/shared/proto/runtime"
	zlog "github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type gRPCServer struct {
	mu         sync.RWMutex
	executions map[string]*Execution
	ctx        context.Context
	dispatcher *client.Dispatcher
	pb.UnimplementedRuntimeServiceServer
}

func NewGrpc(ctx context.Context, server *grpc.Server, dispatcher *client.Dispatcher) *gRPCServer {
	srv := &gRPCServer{
		executions: make(map[string]*Execution),
		dispatcher: dispatcher,
		mu:         sync.RWMutex{},
		ctx:        ctx,
	}

	go srv.startHeartbeatLoop(ctx)
	pb.RegisterRuntimeServiceServer(server, srv)
	zlog.Info().Msg("registered runtime gRPC service")
	return srv
}

func (g *gRPCServer) Run(ctx context.Context, req *pb.RunRequest) (*pb.RunResponse, error) {
	executionID := ids.NewExecutionID()

	exec := &Execution{
		ID:        executionID,
		JobID:     req.JobId,
		AttemptID: req.AttemptId,
		StartedAt: time.Now().UTC(),
	}
	exec.Status.Store(
		int32(pb.ExecutionStatus_EXECUTION_STATUS_PENDING),
	)

	zlog.Info().
		Str("execution_id", executionID).
		Str("job_id", req.JobId).
		Str("attempt_id", req.AttemptId).
		Msg("execution registered")

	g.mu.Lock()
	g.executions[executionID] = exec
	count := len(g.executions)
	g.mu.Unlock()

	zlog.Info().
		Int("active_executions", count).
		Msg("execution registered")

	go g.execute(ctx, exec, req)

	return &pb.RunResponse{
		ExecutionId: executionID,
	}, nil
}

func (g *gRPCServer) startHeartbeatLoop(
	ctx context.Context,
) {
	for {
		if ctx.Err() != nil {
			return
		}

		err := g.heartbeatLoop(ctx)
		if err == nil {
			return
		}

		zlog.Warn().
			Err(err).
			Msg("heartbeat stream disconnected, reconnecting")

		time.Sleep(5 * time.Second)
	}
}

func (g *gRPCServer) heartbeatLoop(
	ctx context.Context,
) error {
	stream, err := g.dispatcher.Heartbeat(ctx)
	if err != nil {
		return err
	}

	zlog.Info().Msg("heartbeat stream established")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			g.mu.RLock()
			for _, exec := range g.executions {
				if pb.ExecutionStatus(exec.Status.Load()) != pb.ExecutionStatus_EXECUTION_STATUS_RUNNING {
					continue
				}

				zlog.Debug().
					Str("execution_id", exec.ID).
					Str("job_id", exec.JobID).
					Int64("pid", exec.PID.Load()).
					Msg("sending execution heartbeat")
				if err := stream.Send(&dispatcherpb.HeartbeatRequest{
					ExecutionId: exec.ID,
					JobId:       exec.JobID,
					AttemptId:   exec.AttemptID,
					Status:      dispatcherpb.ExecutionStatus(exec.Status.Load()),
					Pid:         exec.PID.Load(),
					UnixTime:    time.Now().Unix(),
				}); err != nil {
					g.mu.RUnlock()
					return err
				}
			}
			g.mu.RUnlock()
		}
	}
}
