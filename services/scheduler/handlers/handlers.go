package handlers

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/aliamerj/wardu/shared/database"
	"github.com/aliamerj/wardu/shared/k8s"
	"github.com/aliamerj/wardu/shared/models"
	pb "github.com/aliamerj/wardu/shared/proto/scheduler"
	r "github.com/aliamerj/wardu/shared/rabbitmq"
	"github.com/oklog/ulid/v2"
	zlog "github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type Handler struct {
	db       database.Service
	k8s      *k8s.Client
	rabbitmq *r.RabbitMQ
}

func New(db database.Service, k8s *k8s.Client, rabbitmq *r.RabbitMQ) *Handler {
	h := &Handler{
		db:       db,
		k8s:      k8s,
		rabbitmq: rabbitmq,
	}

	if err := h.createDefualtNamespace(); err != nil {
		zlog.Fatal().Err(err).Msg("failed to bootstrap default namespace")
	}

	return h
}

func (h *Handler) CreateJob(
	ctx context.Context,
	req *pb.CreateJobRequest,
) (string, error) {
	started := time.Now()
	job := models.BuildJobFromProto(req)
	job.ID = newJobID()

	zlog.Info().
		Str("job_id", job.ID).
		Str("namespace", job.Namespace).
		Str("image", req.GetImage()).
		Msg("received job creation request")

	ns, err := h.db.GetNamespaceByName(job.Namespace)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Warn().
				Str("job_id", job.ID).
				Str("namespace", job.Namespace).
				Msg("job namespace not found")
			return "", fmt.Errorf(
				"namespace %s not found, please create the namespace or use default wardu",
				job.Namespace,
			)
		}

		zlog.Error().Err(err).Str("job_id", job.ID).Msg("failed to load namespace from database")
		return "", err
	}

	exists, err := h.k8s.CheckNamespaceByDNS(ctx, ns.DNS)
	if err != nil {
		zlog.Error().Err(err).Str("job_id", job.ID).Str("namespace", ns.Name).Msg("failed to check namespace in kubernetes")
		return "", err
	}

	if !exists {
		zlog.Info().Str("job_id", job.ID).Str("namespace", ns.Name).Msg("namespace missing in kubernetes, creating")

		if _, err := h.k8s.CreateNamespace(ctx, ns.Name, ns); err != nil {
			zlog.Error().Err(err).Str("job_id", job.ID).Str("namespace", ns.Name).Msg("failed to create kubernetes namespace")
			return "", err
		}
	}

	worker, err := h.db.GetWorkerByImage(req.GetImage())
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Error().Err(err).Str("job_id", job.ID).Str("image", req.GetImage()).Msg("failed to load worker from database")
			return "", err
		}

		zlog.Info().
			Str("job_id", job.ID).
			Str("image", req.GetImage()).
			Msg("worker not found, creating new worker")

		worker, err = h.k8s.CreateWorker(ctx, ns, job, req.GetImage())
		if err != nil {
			zlog.Error().Err(err).Str("job_id", job.ID).Str("image", req.GetImage()).Msg("failed to create worker in kubernetes")
			return "", err
		}

		if err := h.db.CreateWorker(worker); err != nil {
			zlog.Error().Err(err).Str("job_id", job.ID).Str("image", req.GetImage()).Msg("failed to persist worker")
			return "", err
		}
	} else {
		exists, err := h.k8s.CheckWorker(ctx, ns.DNS, worker.K8sDeploymentName)
		if err != nil {
			zlog.Error().Err(err).Str("job_id", job.ID).Str("worker", worker.K8sDeploymentName).Msg("failed to check worker in kubernetes")
			return "", err
		}

		if !exists {
			zlog.Info().
				Str("job_id", job.ID).
				Str("worker", worker.K8sDeploymentName).
				Msg("worker deployment missing in kubernetes, recreating")

			newWorker, err := h.k8s.CreateWorker(ctx, ns, job, req.GetImage())
			if err != nil {
				zlog.Error().Err(err).Str("job_id", job.ID).Str("image", req.GetImage()).Msg("failed to recreate worker in kubernetes")
				return "", err
			}

			worker.K8sDeploymentName = newWorker.K8sDeploymentName

			if err := h.db.UpdateWorker(worker); err != nil {
				zlog.Error().Err(err).Str("job_id", job.ID).Str("image", req.GetImage()).Msg("failed to update worker deployment name")
				return "", err
			}
		}
	}

	job.WorkerID = worker.ID

	if err := h.db.CreateJob(job); err != nil {
		zlog.Error().Err(err).Str("job_id", job.ID).Str("worker_id", job.WorkerID).Msg("failed to persist job")
		return "", err
	}

	zlog.Info().
		Str("job_id", job.ID).
		Str("worker_id", job.WorkerID).
		Dur("latency", time.Since(started)).
		Msg("job created successfully")

	if job.Autorun {
		if err := h.rabbitmq.PublishJob(ctx, r.JobMessage{
			JobID:    job.ID,
			Image:    req.GetImage(),
			Priority: req.GetPriority(),
			Attempt:  1,
		}); err != nil {
			zlog.Error().Err(err).Str("job_id", job.ID).Msg("failed to publish job to RabbitMQ")
			return "", err
		}
		zlog.Info().Str("job_id", job.ID).Msg("job successfully queued")
	}

	return job.ID, nil
}

func newJobID() string {
	return ulid.MustNew(
		ulid.Timestamp(time.Now()),
		ulid.Monotonic(rand.Reader, 0),
	).String()
}

func (h *Handler) createDefualtNamespace() error {
	ctx := context.Background()
	nss, err := h.db.GetAllNamespaces()
	if err != nil {
		return err
	}

	hasWardu := false
	for _, ns := range nss {
		if ns.Name == "wardu" {
			hasWardu = true
		}

		exist, err := h.k8s.CheckNamespaceByDNS(ctx, ns.DNS)
		if err != nil {
			return err
		}
		if !exist {
			zlog.Warn().Str("namespace", ns.Name).Str("dns", ns.DNS).Msg("namespace missing in kubernetes, recreating")

			if _, err := h.k8s.CreateNamespace(ctx, ns.Name, ns); err != nil {
				return err
			}
		}
	}

	if hasWardu {
		return nil
	}

	zlog.Info().Msg("bootstrapping default wardu namespace")
	ns, err := h.k8s.CreateNamespace(ctx, "wardu", nil)
	if err != nil {
		return err
	}

	if err := h.db.CreateNamespace(ns); err != nil {
		return err
	}

	zlog.Info().Str("namespace", ns.Name).Msg("default namespace bootstrapped")
	return nil
}
