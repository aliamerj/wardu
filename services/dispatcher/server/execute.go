package server

import (
	"time"

	"github.com/aliamerj/wardu/shared/ids"
	"github.com/aliamerj/wardu/shared/models"
	r "github.com/aliamerj/wardu/shared/rabbitmq"
	zlog "github.com/rs/zerolog/log"
)

func (g *gRPCServer) executeJob(
	jm r.JobMessage,
) error {
	zlog.Info().
		Str("job_id", jm.JobID).
		Int("attempt", jm.Attempt).
		Msg("starting job execution")

	job, err := g.db.GetJobForExecution(jm.JobID)
	if err != nil {
		zlog.Error().
			Err(err).
			Str("job_id", jm.JobID).
			Msg("failed to load job")

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
		zlog.Error().
			Err(err).
			Str("job_id", job.ID).
			Str("attempt_id", attempt.ID).
			Msg("failed to create job attempt")

		return err
	}
	var erro error

	defer func() {
		if erro != nil {
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
	}()

	zlog.Info().
		Str("job_id", job.ID).
		Str("attempt_id", attempt.ID).
		Int("attempt", jm.Attempt).
		Msg("job attempt created")

	if err := g.k8s.ScaleWorker(g.ctx, job, 1, true); err != nil {
		zlog.Error().
			Err(err).
			Str("job_id", job.ID).
			Str("attempt_id", attempt.ID).
			Msg("failed to wake up worker")
		return err
	}

	// TODO:
	// 1. Run Job
	// 2. save result in job attempt and update status

	return nil
}
