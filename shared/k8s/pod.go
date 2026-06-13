package k8s

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/aliamerj/wardu/shared/models"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) SelectPod(
	ctx context.Context,
	job *models.Job,
) (*corev1.Pod, error) {
	var ready []corev1.Pod

	pods, err := c.k8s.
		CoreV1().
		Pods(job.Worker.Namespace.DNS).
		List(
			ctx,
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf(
					"worker=%s",
					job.Worker.K8sDeploymentName,
				),
			},
		)
	if err != nil {
		return nil, err
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase != corev1.PodRunning {
			continue
		}

		for _, cond := range pod.Status.Conditions {
			if cond.Type == corev1.PodReady &&
				cond.Status == corev1.ConditionTrue {

				ready = append(ready, pod)
				break
			}
		}
	}

	if len(ready) == 0 {
		return nil, fmt.Errorf("no ready worker pods")
	}
	// TODO: Repace with Pick Least Busy Pod
	pod := ready[rand.Intn(len(ready))]

	return &pod, nil
}
