package handlers

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aliamerj/wardu/shared/database"
	"github.com/aliamerj/wardu/shared/k8s"
	"github.com/aliamerj/wardu/shared/models"
	pb "github.com/aliamerj/wardu/shared/proto/scheduler"
	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type Handler struct {
	db  database.Service
	k8s *k8s.Client
}

func New(db database.Service, k8s *k8s.Client) *Handler {
	h := &Handler{
		db:  db,
		k8s: k8s,
	}

	if err := h.createDefualtNamespace(); err != nil {
		log.Fatalf("failed to create defualt Namespace: %s", err.Error())
	}

	return h
}

func (h *Handler) CreateJob(
	ctx context.Context,
	req *pb.CreateJobRequest,
) (string, error) {
	job := models.BuildJobFromProto(req)
	job.ID = newJobID()

	ns, err := h.db.GetNamespaceByName(job.Namespace)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", fmt.Errorf(
				"namespace %s not found, please create the namespace or use default wardu",
				job.Namespace,
			)
		}

		return "", err
	}

	exists, err := h.k8s.CheckNamespaceByDNS(ctx, ns.DNS)
	if err != nil {
		return "", err
	}

	if !exists {
		log.Printf("namespace %s missing in k8s", ns.Name)

		if _, err := h.k8s.CreateNamespace(
			ctx,
			ns.Name,
			ns,
		); err != nil {
			return "", err
		}
	}

	worker, err := h.db.GetWorkerByImage(req.GetImage())
	if err != nil {

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return "", err
		}

		log.Printf(
			"worker for image %s not found, creating",
			req.GetImage(),
		)

		worker, err = h.k8s.CreateWorker(
			ctx,
			ns,
			job,
			req.GetImage(),
		)
		if err != nil {
			return "", err
		}

		if err := h.db.CreateWorker(worker); err != nil {
			return "", err
		}
	} else {

		exists, err := h.k8s.CheckWorker(
			ctx,
			ns.DNS,
			worker.K8sDeploymentName,
		)
		if err != nil {
			return "", err
		}

		if !exists {

			log.Printf(
				"worker deployment %s missing in k8s, recreating",
				worker.K8sDeploymentName,
			)

			newWorker, err := h.k8s.CreateWorker(
				ctx,
				ns,
				job,
				req.GetImage(),
			)
			if err != nil {
				return "", err
			}

			// keep existing DB record
			worker.K8sDeploymentName = newWorker.K8sDeploymentName

			if err := h.db.UpdateWorker(worker); err != nil {
				return "", err
			}
		}
	}

	job.WorkerID = worker.ID

	if err := h.db.CreateJob(job); err != nil {
		return "", err
	}

	// TODO:
	// if job.Autorun {
	//     publish to rabbitmq
	// }

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
			log.Printf(
				"namespace %s missing in k8s\n",
				ns.DNS,
			)
			log.Printf("creating namespace %s in k8s", ns.Name)

			if _, err := h.k8s.CreateNamespace(ctx, ns.Name, ns); err != nil {
				return err
			}
		}
	}

	if hasWardu {
		return nil
	}

	ns, err := h.k8s.CreateNamespace(ctx, "wardu", nil)
	if err != nil {
		return err
	}

	return h.db.CreateNamespace(ns)
}
