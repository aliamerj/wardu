package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aliamerj/wardu/shared/database"
	"github.com/aliamerj/wardu/shared/ids"
	"github.com/aliamerj/wardu/shared/k8s"
	"github.com/aliamerj/wardu/shared/models"
	pb "github.com/aliamerj/wardu/shared/proto/scheduler"
	r "github.com/aliamerj/wardu/shared/rabbitmq"
	zlog "github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type gRPCServer struct {
	ctx      context.Context
	db       database.Service
	k8s      *k8s.Client
	rabbitmq *r.RabbitMQ
	pb.UnimplementedSchedulerServiceServer
}

func NewGrpc(ctx context.Context, server *grpc.Server, rabbitmq *r.RabbitMQ) *gRPCServer {
	srv := &gRPCServer{
		ctx:      ctx,
		db:       database.New(),
		k8s:      k8s.New(),
		rabbitmq: rabbitmq,
	}

	if err := srv.createDefualtNamespace(); err != nil {
		zlog.Fatal().Err(err).Msg("failed to bootstrap default namespace")
	}

	pb.RegisterSchedulerServiceServer(server, srv)
	zlog.Info().Msg("registered scheduler gRPC service")
	return srv
}

func (g *gRPCServer) CreateJob(ctx context.Context, req *pb.CreateJobRequest) (*pb.CreateJobResponse, error) {
	started := time.Now()
	job := models.BuildJobFromProto(req)
	job.ID = ids.NewJobID()

	zlog.Info().
		Str("job_id", job.ID).
		Str("namespace", job.Namespace).
		Str("image", req.GetImage()).
		Msg("received job creation request")

	ns, err := g.db.GetNamespaceByName(job.Namespace)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Warn().
				Str("job_id", job.ID).
				Str("namespace", job.Namespace).
				Msg("job namespace not found")
			return nil, fmt.Errorf(
				"namespace %s not found, please create the namespace or use default wardu",
				job.Namespace,
			)
		}

		zlog.Error().Err(err).Str("job_id", job.ID).Msg("failed to load namespace from database")
		return nil, err
	}

	exists, err := g.k8s.CheckNamespaceByDNS(ctx, ns.DNS)
	if err != nil {
		zlog.Error().Err(err).Str("job_id", job.ID).Str("namespace", ns.Name).Msg("failed to check namespace in kubernetes")
		return nil, err
	}

	if !exists {
		zlog.Info().Str("job_id", job.ID).Str("namespace", ns.Name).Msg("namespace missing in kubernetes, creating")

		if _, err := g.k8s.CreateNamespace(ctx, ns.Name, ns); err != nil {
			zlog.Error().Err(err).Str("job_id", job.ID).Str("namespace", ns.Name).Msg("failed to create kubernetes namespace")
			return nil, err
		}
	}

	worker, err := g.db.GetWorkerByImage(req.GetImage())
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Error().Err(err).Str("job_id", job.ID).Str("image", req.GetImage()).Msg("failed to load worker from database")
			return nil, err
		}

		zlog.Info().
			Str("job_id", job.ID).
			Str("image", req.GetImage()).
			Msg("worker not found, creating new worker")

		worker, err = g.k8s.CreateWorker(ctx, ns, job, req.GetImage())
		if err != nil {
			zlog.Error().Err(err).Str("job_id", job.ID).Str("image", req.GetImage()).Msg("failed to create worker in kubernetes")
			return nil, err
		}

		if err := g.db.CreateWorker(worker); err != nil {
			zlog.Error().Err(err).Str("job_id", job.ID).Str("image", req.GetImage()).Msg("failed to persist worker")
			return nil, err
		}
	} else {
		exists, err := g.k8s.CheckWorker(ctx, ns.DNS, worker.K8sDeploymentName)
		if err != nil {
			zlog.Error().Err(err).Str("job_id", job.ID).Str("worker", worker.K8sDeploymentName).Msg("failed to check worker in kubernetes")
			return nil, err
		}

		if !exists {
			zlog.Info().
				Str("job_id", job.ID).
				Str("worker", worker.K8sDeploymentName).
				Msg("worker deployment missing in kubernetes, recreating")

			newWorker, err := g.k8s.CreateWorker(ctx, ns, job, req.GetImage())
			if err != nil {
				zlog.Error().Err(err).Str("job_id", job.ID).Str("image", req.GetImage()).Msg("failed to recreate worker in kubernetes")
				return nil, err
			}

			worker.K8sDeploymentName = newWorker.K8sDeploymentName

			if err := g.db.UpdateWorker(worker); err != nil {
				zlog.Error().Err(err).Str("job_id", job.ID).Str("image", req.GetImage()).Msg("failed to update worker deployment name")
				return nil, err
			}
		}
	}

	job.WorkerID = worker.ID

	if err := g.db.CreateJob(job); err != nil {
		zlog.Error().Err(err).Str("job_id", job.ID).Str("worker_id", job.WorkerID).Msg("failed to persist job")
		return nil, err
	}

	zlog.Info().
		Str("job_id", job.ID).
		Str("worker_id", job.WorkerID).
		Dur("latency", time.Since(started)).
		Msg("job created successfully")

	if job.Autorun {
		if err := g.rabbitmq.PublishJob(ctx, r.JobMessage{
			JobID:    job.ID,
			Priority: req.GetPriority(),
			Attempt:  1,
		}); err != nil {
			zlog.Error().Err(err).Str("job_id", job.ID).Msg("failed to publish job to RabbitMQ")
			return nil, err
		}
		zlog.Info().Str("job_id", job.ID).Msg("job successfully queued")
	}

	return &pb.CreateJobResponse{
		JobId: job.ID,
	}, nil
}

func (g *gRPCServer) createDefualtNamespace() error {
	nss, err := g.db.GetAllNamespaces()
	if err != nil {
		return err
	}

	hasWardu := false
	for _, ns := range nss {
		if ns.Name == "wardu" {
			hasWardu = true
		}

		exist, err := g.k8s.CheckNamespaceByDNS(g.ctx, ns.DNS)
		if err != nil {
			return err
		}
		if !exist {
			zlog.Warn().Str("namespace", ns.Name).Str("dns", ns.DNS).Msg("namespace missing in kubernetes, recreating")

			if _, err := g.k8s.CreateNamespace(g.ctx, ns.Name, ns); err != nil {
				return err
			}
		}
	}

	if hasWardu {
		return nil
	}

	zlog.Info().Msg("bootstrapping default wardu namespace")
	ns, err := g.k8s.CreateNamespace(g.ctx, "wardu", nil)
	if err != nil {
		return err
	}

	if err := g.db.CreateNamespace(ns); err != nil {
		return err
	}

	zlog.Info().Str("namespace", ns.Name).Msg("default namespace bootstrapped")
	return nil
}
