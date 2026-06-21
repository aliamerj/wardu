package server

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync/atomic"
	"syscall"
	"time"

	pb "github.com/aliamerj/wardu/shared/proto/runtime"
	zlog "github.com/rs/zerolog/log"
)

type Execution struct {
	ID string

	JobID     string
	AttemptID string

	Status atomic.Int32
	PID    atomic.Int64

	StartedAt  time.Time
	FinishedAt *time.Time

	Result []byte
	Error  string
}

func (g *gRPCServer) execute(ctx context.Context, exec *Execution, req *pb.RunRequest) {
	defer func() {
		g.mu.Lock()
		delete(g.executions, exec.ID)
		count := len(g.executions)
		g.mu.Unlock()

		zlog.Info().
			Int("active_executions", count).
			Str("execution_id", exec.ID).
			Msg("execution removed from runtime")
	}()

	exec.Status.Store(
		int32(pb.ExecutionStatus_EXECUTION_STATUS_RUNNING),
	)
	exec.StartedAt = time.Now().UTC()

	zlog.Info().
		Str("job_id", req.JobId).
		Strs("entrypoint", req.Entrypoint).
		Msg("starting user process")

	cmd, stdout, stderr, err := startProcess(req)
	if err != nil {
		exec.Status.Store(
			int32(pb.ExecutionStatus_EXECUTION_STATUS_FAILED),
		)
		exec.Error = err.Error()
		zlog.Error().
			Err(err).
			Str("job_id", req.JobId).
			Msg("failed to start process")
		return
	}

	exec.PID.Store(
		int64(cmd.Process.Pid),
	)

	zlog.Info().
		Str("job_id", req.JobId).
		Int("pid", cmd.Process.Pid).
		Msg("process started")

	go collectLogs(
		exec,
		"stdout",
		stdout,
	)

	go collectLogs(
		exec,
		"stderr",
		stderr,
	)

	timeout := time.Duration(req.TimeoutSeconds) * time.Second
	waitCtx := ctx

	if timeout > 0 {
		var cancel context.CancelFunc

		waitCtx, cancel = context.WithTimeout(
			ctx,
			timeout,
		)

		defer cancel()
	}

	done := make(chan error, 1)

	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		now := time.Now().UTC()
		exec.FinishedAt = &now
		if err != nil {
			exec.Status.Store(
				int32(pb.ExecutionStatus_EXECUTION_STATUS_FAILED),
			)
			exec.Error = err.Error()

			zlog.Error().
				Err(err).
				Str("job_id", req.JobId).
				Int("pid", cmd.Process.Pid).
				Msg("process failed")
			return
		}
		exec.Status.Store(
			int32(pb.ExecutionStatus_EXECUTION_STATUS_SUCCEEDED),
		)
		zlog.Info().
			Str("job_id", req.JobId).
			Int("pid", cmd.Process.Pid).
			Msg("process completed")

	case <-waitCtx.Done():
		_ = cmd.Process.Signal(syscall.SIGTERM)

		select {
		case <-time.After(10 * time.Second):
			_ = cmd.Process.Kill()
		case <-done:
		}

		now := time.Now().UTC()
		exec.FinishedAt = &now

		exec.Status.Store(
			int32(pb.ExecutionStatus_EXECUTION_STATUS_FAILED),
		)
		exec.Error = waitCtx.Err().Error()

		zlog.Warn().
			Str("job_id", req.JobId).
			Int("pid", cmd.Process.Pid).
			Msg("process terminated")

	}
}

func startProcess(
	req *pb.RunRequest,
) (*exec.Cmd, io.ReadCloser, io.ReadCloser, error) {
	cmd := exec.Command(
		req.Entrypoint[0],
		req.Entrypoint[1:]...,
	)

	cmd.Env = append(
		os.Environ(),
		fmt.Sprintf(
			"WARDU_JOB_ID=%s",
			req.JobId,
		),
		fmt.Sprintf(
			"WARDU_ATTEMPT_ID=%s",
			req.AttemptId,
		),
		fmt.Sprintf(
			"WARDU_PAYLOAD=%s",
			base64.StdEncoding.EncodeToString(
				req.Payload,
			),
		),
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, nil, err
	}

	return cmd, stdout, stderr, nil
}

func collectLogs(
	execRec *Execution,
	stream string,
	r io.Reader,
) {
	scanner := bufio.NewScanner(r)
	if err := scanner.Err(); err != nil {
		zlog.Error().
			Err(err).
			Str("job_id", execRec.JobID).
			Str("stream", stream).
			Msg("log stream failed")
	}

	for scanner.Scan() {

		line := scanner.Text()

		zlog.Info().
			Str("job_id", execRec.JobID).
			Str("stream", stream).
			Msg(line)

		// TODO later:
		// publish log event
		// persist log
		// websocket fanout
	}

	if err := scanner.Err(); err != nil {
		zlog.Error().
			Err(err).
			Str("job_id", execRec.JobID).
			Str("stream", stream).
			Msg("log stream error")
	}
}
