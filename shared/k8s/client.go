package k8s

import (
	"log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Client struct {
	k8s *kubernetes.Clientset
}

func New() *Client {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("failed to start kubernete: %s", err.Error())
		return nil
	}

	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("failed to start kubernete: %s", err.Error())
		return nil
	}
	return &Client{
		k8s: cs,
	}
}
