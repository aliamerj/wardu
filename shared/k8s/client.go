package k8s

import (
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Client struct {
	k8s *kubernetes.Clientset
}

func New() *Client {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load in-cluster kubernetes configuration")
	}

	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create kubernetes clientset")
	}

	log.Info().Msg("kubernetes client initialized")
	return &Client{k8s: cs}
}
