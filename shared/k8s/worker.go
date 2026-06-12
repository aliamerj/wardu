package k8s

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aliamerj/wardu/shared/models"
	"github.com/google/uuid"
	zlog "github.com/rs/zerolog/log"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) CreateWorker(ctx context.Context, ns *models.Namespace, job *models.Job, image string) (*models.Worker, error) {
	replicas := int32(0)
	deploymentName := generateDeploymentName(image, ns.DNS)

	deployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: ns.DNS,
			Labels: map[string]string{
				"managed-by": "wardu",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"managed-by": "wardu",
					"worker":     deploymentName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"managed-by": "wardu",
						"worker":     deploymentName,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "worker",
							Image:   image,
							Command: job.Entrypoint,
						},
					},
				},
			},
		},
	}

	if _, err := c.k8s.
		AppsV1().
		Deployments(ns.DNS).
		Create(ctx, &deployment, metav1.CreateOptions{}); err != nil {
		return nil, err
	}

	worker := models.Worker{
		ID:                 uuid.NewString(),
		Image:              image,
		NamespaceID:        ns.ID,
		K8sDeploymentName:  deploymentName,
		IdleTimeoutSeconds: int(job.IdleTimeoutSeconds),
		MaxReplicas:        ns.MaxWorkers,
	}

	return &worker, nil
}

func (c *Client) CheckWorker(ctx context.Context, namespace, k8sDeploymentName string) (bool, error) {
	_, err := c.k8s.AppsV1().Deployments(namespace).Get(ctx, k8sDeploymentName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func sanitizeImageName(image string) string {
	image = strings.ToLower(image)
	image = strings.ReplaceAll(image, "/", "-")
	image = strings.ReplaceAll(image, ":", "-")
	image = strings.ReplaceAll(image, "_", "-")
	return image
}

func generateDeploymentName(image, namespace string) string {
	imageName := sanitizeImageName(image)

	name := "worker-" + namespace + "-" + imageName

	// Kubernetes safety: max 63 chars for DNS label (safe for Deployment name too)
	if len(name) > 63 {
		name = name[:63]
	}

	// avoid trailing hyphen after trimming
	name = strings.Trim(name, "-")

	return name
}

func (c *Client) ScaleWorker(ctx context.Context, job *models.Job, replicas int32, isWakeUpCall bool) error {
	if replicas > int32(job.Worker.MaxReplicas) {
		zlog.Error().
			Str("job_id", job.ID).
			Str("worker", job.Worker.K8sDeploymentName).
			Int32("requested", replicas).
			Int("max_replicas", job.Worker.MaxReplicas).
			Msg("requested replicas exceed worker limit")

		return fmt.Errorf(
			"requested replicas %d exceeds max replicas %d",
			replicas,
			job.Worker.MaxReplicas,
		)
	}

	scaleObj, err := c.getWorker(
		ctx,
		job.Worker.Namespace.DNS,
		job.Worker.K8sDeploymentName,
	)
	if err != nil {
		if apierrors.IsNotFound(err) {
			zlog.Warn().
				Str("job_id", job.ID).
				Str("worker", job.Worker.K8sDeploymentName).
				Msg("worker deployment missing, starting recovery")

			if err := c.recoverWorker(ctx, job); err != nil {
				zlog.Error().
					Err(err).
					Str("job_id", job.ID).
					Str("worker", job.Worker.K8sDeploymentName).
					Msg("worker recovery failed")

				return err
			}

			scaleObj, err = c.getWorker(
				ctx,
				job.Worker.Namespace.DNS,
				job.Worker.K8sDeploymentName,
			)
			if err != nil {
				zlog.Error().
					Err(err).
					Str("job_id", job.ID).
					Str("worker", job.Worker.K8sDeploymentName).
					Msg("failed to fetch recovered worker deployment")

				return err
			}

			zlog.Info().
				Str("deployment", job.Worker.K8sDeploymentName).
				Msg("worker recovered successfully")

		} else {
			zlog.Error().
				Err(err).
				Str("job_id", job.ID).
				Str("worker", job.Worker.K8sDeploymentName).
				Msg("failed to fetch worker deployment")

			return err
		}
	}

	current := scaleObj.Spec.Replicas

	if isWakeUpCall && current != 0 {
		zlog.Debug().
			Str("job_id", job.ID).
			Str("worker", job.Worker.K8sDeploymentName).
			Int32("replicas", current).
			Msg("worker already awake, skipping scale")

		return nil
	}

	if current == replicas {
		zlog.Debug().
			Str("job_id", job.ID).
			Str("worker", job.Worker.K8sDeploymentName).
			Int32("replicas", replicas).
			Msg("worker already at requested replica count")

		return nil
	}

	if replicas > current {
		zlog.Info().
			Str("worker", job.Worker.K8sDeploymentName).
			Int32("from", current).
			Int32("to", replicas).
			Msg("scaling worker up")
	}

	if replicas < current {
		zlog.Info().
			Str("worker", job.Worker.K8sDeploymentName).
			Int32("from", current).
			Int32("to", replicas).
			Msg("scaling worker down")
	}

	scaleObj.Spec.Replicas = replicas

	if _, err = c.k8s.
		AppsV1().
		Deployments(job.Worker.Namespace.DNS).
		UpdateScale(
			ctx,
			job.Worker.K8sDeploymentName,
			scaleObj,
			metav1.UpdateOptions{},
		); err != nil {

		zlog.Error().
			Err(err).
			Str("job_id", job.ID).
			Str("worker", job.Worker.K8sDeploymentName).
			Int32("target_replicas", replicas).
			Msg("failed to update worker deployment replicas")

		return err
	}

	zlog.Info().
		Str("job_id", job.ID).
		Str("worker", job.Worker.K8sDeploymentName).
		Int32("replicas", replicas).
		Msg("worker scaled successfully")

	if current == 0 && replicas > 0 {
		return c.WaitWorkerReady(ctx, job, time.Minute)
	}

	return nil
}

func (c *Client) WaitWorkerReady(
	ctx context.Context,
	job *models.Job,
	timeout time.Duration,
) error {
	zlog.Info().
		Str("job_id", job.ID).
		Str("worker", job.Worker.K8sDeploymentName).
		Dur("timeout", timeout).
		Msg("waiting for worker to become ready")

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	deployment, err := c.getWorker(
		timeoutCtx,
		job.Worker.Namespace.DNS,
		job.Worker.K8sDeploymentName,
	)
	if err != nil {
		zlog.Error().
			Err(err).
			Str("job_id", job.ID).
			Str("worker", job.Worker.K8sDeploymentName).
			Msg("failed to get deployment while waiting for readiness")

		return err
	}

	if deployment.Status.Replicas > 0 {
		zlog.Info().
			Str("job_id", job.ID).
			Str("worker", job.Worker.K8sDeploymentName).
			Int32("ready_replicas", deployment.Status.Replicas).
			Msg("worker already ready")

		return nil
	}

	watcher, err := c.k8s.
		AppsV1().
		Deployments(job.Worker.Namespace.DNS).
		Watch(
			timeoutCtx,
			metav1.ListOptions{
				FieldSelector: fmt.Sprintf(
					"metadata.name=%s",
					job.Worker.K8sDeploymentName,
				),
			},
		)
	if err != nil {
		zlog.Error().
			Err(err).
			Str("job_id", job.ID).
			Str("worker", job.Worker.K8sDeploymentName).
			Msg("failed to create deployment watcher")

		return err
	}

	defer watcher.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			zlog.Error().
				Str("job_id", job.ID).
				Str("worker", job.Worker.K8sDeploymentName).
				Dur("timeout", timeout).
				Msg("timed out waiting for worker readiness")

			return timeoutCtx.Err()

		case event, ok := <-watcher.ResultChan():
			if !ok {
				zlog.Error().
					Str("job_id", job.ID).
					Str("worker", job.Worker.K8sDeploymentName).
					Msg("deployment watch channel closed")

				return fmt.Errorf("deployment watch closed")
			}

			d, ok := event.Object.(*appsv1.Deployment)
			if !ok {
				continue
			}

			zlog.Debug().
				Str("job_id", job.ID).
				Str("worker", d.Name).
				Int32("desired", func() int32 {
					if d.Spec.Replicas == nil {
						return 0
					}
					return *d.Spec.Replicas
				}()).
				Int32("ready", d.Status.ReadyReplicas).
				Int32("available", d.Status.AvailableReplicas).
				Msg("deployment state changed")

			if d.Status.ReadyReplicas > 0 {
				zlog.Info().
					Str("job_id", job.ID).
					Str("worker", d.Name).
					Int32("ready_replicas", d.Status.ReadyReplicas).
					Msg("worker is ready")

				return nil
			}
		}
	}
}

func (c *Client) getWorker(ctx context.Context, namespace, K8sDeploymentName string) (*autoscalingv1.Scale, error) {
	return c.k8s.
		AppsV1().
		Deployments(namespace).
		GetScale(
			ctx,
			K8sDeploymentName,
			metav1.GetOptions{},
		)
}

func (c *Client) recoverWorker(
	ctx context.Context,
	job *models.Job,
) error {
	ns := &job.Worker.Namespace

	// recover namespace
	zlog.Info().
		Str("namespace", job.Worker.Namespace.DNS).
		Msg("recovering namespace")
	if _, err := c.CreateNamespace(
		ctx,
		ns.Name,
		ns,
	); err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	// recover deployment
	zlog.Info().
		Str("deployment", job.Worker.K8sDeploymentName).
		Msg("recovering worker deployment")
	if _, err := c.CreateWorker(
		ctx,
		ns,
		job,
		job.Worker.Image,
	); err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	return nil
}
