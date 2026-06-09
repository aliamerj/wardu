package k8s

import (
	"context"
	"strings"

	"github.com/aliamerj/wardu/shared/models"
	"github.com/google/uuid"
	appsv1 "k8s.io/api/apps/v1"
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
