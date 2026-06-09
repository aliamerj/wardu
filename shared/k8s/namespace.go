package k8s

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/aliamerj/wardu/shared/models"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) CreateNamespace(ctx context.Context, name string, ops *models.Namespace) (*models.Namespace, error) {
	ns := buildDefaultNamespace(name)

	if ops != nil {
		if ops.MaxWorkers > 0 {
			ns.MaxWorkers = ops.MaxWorkers
		}

		if ops.MaxConcurrentJobs > 0 {
			ns.MaxConcurrentJobs = ops.MaxConcurrentJobs
		}

		if ops.MaxPods > 0 {
			ns.MaxPods = ops.MaxPods
		}

		if ops.CPURequestMilli > 0 {
			ns.CPURequestMilli = ops.CPURequestMilli
		}

		if ops.CPULimitMilli > 0 {
			ns.CPULimitMilli = ops.CPULimitMilli
		}

		if ops.MemoryRequestMB > 0 {
			ns.MemoryRequestMB = ops.MemoryRequestMB
		}
		if ops.MemoryLimitMB > 0 {
			ns.MemoryLimitMB = ops.MemoryLimitMB
		}
	}

	if err := c.createNamespace(ctx, ns.DNS); err != nil && !apierrors.IsAlreadyExists(err) {
		return nil, err
	}

	if err := c.createResourceQuota(ctx, ns); err != nil && !apierrors.IsAlreadyExists(err) {
		return nil, err
	}

	if err := c.createLimitRange(ctx, ns); err != nil && !apierrors.IsAlreadyExists(err) {
		return nil, err
	}

	return ns, nil
}

func (c *Client) CheckNamespaceByDNS(ctx context.Context, dns string) (bool, error) {
	_, err := c.k8s.CoreV1().Namespaces().Get(ctx, dns, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (c *Client) createLimitRange(ctx context.Context, ns *models.Namespace) error {
	limitRange := &corev1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-limits", ns.DNS),
			Namespace: ns.DNS,
			Labels: map[string]string{
				"managed-by": "wardu",
			},
		},
		Spec: corev1.LimitRangeSpec{
			Limits: []corev1.LimitRangeItem{
				{
					Type: corev1.LimitTypeContainer,

					DefaultRequest: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse(
							fmt.Sprintf("%dm", ns.CPURequestMilli),
						),
						corev1.ResourceMemory: resource.MustParse(
							fmt.Sprintf("%dMi", ns.MemoryRequestMB),
						),
					},

					Default: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse(
							fmt.Sprintf("%dm", ns.CPULimitMilli),
						),
						corev1.ResourceMemory: resource.MustParse(
							fmt.Sprintf("%dMi", ns.MemoryLimitMB),
						),
					},
				},
			},
		},
	}
	_, err := c.k8s.
		CoreV1().
		LimitRanges(ns.DNS).
		Create(ctx, limitRange, metav1.CreateOptions{})

	return err
}

func (c *Client) createNamespace(ctx context.Context, name string) error {
	namespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"managed-by": "wardu",
			},
		},
	}

	_, err := c.k8s.
		CoreV1().
		Namespaces().
		Create(ctx, &namespace, metav1.CreateOptions{})

	return err
}

func (c *Client) createResourceQuota(ctx context.Context, ns *models.Namespace) error {
	resourceQuota := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-qouta", ns.DNS),
			Namespace: ns.DNS,
			Labels: map[string]string{
				"managed-by": "wardu",
			},
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{
				corev1.ResourcePods: resource.MustParse(strconv.Itoa(ns.MaxPods)),
				corev1.ResourceRequestsCPU: resource.MustParse(
					fmt.Sprintf("%dm", ns.CPURequestMilli),
				),

				corev1.ResourceLimitsCPU: resource.MustParse(
					fmt.Sprintf("%dm", ns.CPULimitMilli),
				),

				corev1.ResourceRequestsMemory: resource.MustParse(
					fmt.Sprintf("%dMi", ns.MemoryRequestMB),
				),

				corev1.ResourceLimitsMemory: resource.MustParse(
					fmt.Sprintf("%dMi", ns.MemoryLimitMB),
				),
			},
		},
	}

	_, err := c.k8s.
		CoreV1().
		ResourceQuotas(ns.DNS).
		Create(ctx, resourceQuota, metav1.CreateOptions{})

	return err
}

func buildDefaultNamespace(name string) *models.Namespace {
	return &models.Namespace{
		Name: name,
		DNS:  generateDNSName(name),

		MaxWorkers:        10,
		MaxConcurrentJobs: 50,
		MaxPods:           20,

		CPURequestMilli: 100,
		CPULimitMilli:   500,

		MemoryRequestMB: 256,
		MemoryLimitMB:   512,
	}
}

func generateDNSName(name string) string {
	name = strings.ToLower(name)

	re := regexp.MustCompile(`[^a-z0-9-]+`)
	name = re.ReplaceAllString(name, "-")

	name = strings.Trim(name, "-")

	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}

	if len(name) > 63 {
		name = name[:63]
	}

	return name
}
