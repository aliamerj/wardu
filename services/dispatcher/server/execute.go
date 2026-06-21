package server

import (
	"context"
	"fmt"
	"net"
	"time"

	client "github.com/aliamerj/wardu/services/dispatcher/runtime_client"
	"github.com/aliamerj/wardu/shared/env"
	"github.com/aliamerj/wardu/shared/ids"
	"github.com/aliamerj/wardu/shared/models"
	pbD "github.com/aliamerj/wardu/shared/proto/dispatcher"
	pb "github.com/aliamerj/wardu/shared/proto/runtime"
	r "github.com/aliamerj/wardu/shared/rabbitmq"
	zlog "github.com/rs/zerolog/log"
)

var RUNTIME_PORT = env.GetInt("RUNTIME_GRPC_PORT", 8083)

func (g *gRPCServer) executeJob(
	jm r.JobMessage,
) error {
	ctx, cancel := context.WithCancel(g.ctx)
	defer cancel()

	zlog.Info().
		Str("job_id", jm.JobID).
		Int("attempt", jm.Attempt).
		Msg("starting job execution")

	job, err := g.db.GetJobForExecution(jm.JobID)
	if err != nil {
		return err
	}

	if jm.Overrides != nil {
		job.ApplyOverrides(jm.Overrides)

		zlog.Info().
			Str("job_id", job.ID).
			Msg("job overrides applied")
	}

	attempt := &models.JobAttempt{
		ID:        ids.NewAttemptID(),
		JobID:     job.ID,
		Attempt:   jm.Attempt,
		Status:    models.AttemptStatusRunning,
		StartedAt: time.Now().UTC(),
	}

	if err := g.db.CreateJobAttempt(attempt); err != nil {
		return err
	}

	zlog.Info().
		Str("job_id", job.ID).
		Str("attempt_id", attempt.ID).
		Msg("job attempt created")

	if err := g.k8s.ScaleWorker(
		ctx,
		job,
		1,
		true,
	); err != nil {
		g.failAttempt(attempt, err)
		return err
	}

	pod, err := g.k8s.SelectPod(
		ctx,
		job,
	)
	if err != nil {
		g.failAttempt(attempt, err)
		return err
	}

	zlog.Info().
		Str("job_id", job.ID).
		Str("pod", pod.Name).
		Msg("worker pod selected")

	runtimeAddr := net.JoinHostPort(
		pod.Status.PodIP,
		fmt.Sprintf(":%d", RUNTIME_PORT),
	)
	runtime, err := client.NewRuntime(runtimeAddr)
	if err != nil {
		return err
	}
	defer runtime.Close()

	resp, err := runtime.Run(ctx, &pb.RunRequest{
		JobId:          job.ID,
		AttemptId:      attempt.ID,
		Entrypoint:     job.Entrypoint,
		Payload:        job.Payload,
		TimeoutSeconds: job.TimeoutSeconds,
	})
	if err != nil {
		g.failAttempt(attempt, err)

		zlog.Error().
			Err(err).
			Str("job_id", job.ID).
			Str("attempt_id", attempt.ID).
			Msg("failed to start runtime execution")

		return err
	}

	execution := &RuntimeExecution{
		ExecutionID: resp.GetExecutionId(),
		JobID:       job.ID,
		AttemptID:   attempt.ID,
		PodName:     pod.Name,
		Status:      pbD.ExecutionStatus_EXECUTION_STATUS_PENDING,
	}

	g.mu.Lock()
	g.executions[execution.ExecutionID] = execution
	g.mu.Unlock()

	zlog.Info().
		Str("execution_id", execution.ExecutionID).
		Str("job_id", execution.JobID).
		Str("attempt_id", execution.AttemptID).
		Str("pod", execution.PodName).
		Msg("execution registered")

	return nil
}

func (g *gRPCServer) failAttempt(
	attempt *models.JobAttempt,
	err error,
) {
	attempt.Status = models.AttemptStatusFailed
	attempt.Error = err.Error()
	attempt.FinishedAt = time.Now().UTC()

	if _, e := g.db.UpdateJobAttempt(attempt); e != nil {
		zlog.Error().
			Err(e).
			Str("attempt_id", attempt.ID).
			Msg("failed to update attempt")
	}
}
