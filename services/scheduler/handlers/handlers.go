package handlers

import (
	"context"
	"log"

	"github.com/aliamerj/wardu/shared/database"
	"github.com/aliamerj/wardu/shared/k8s"
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

func (h *Handler) createDefualtNamespace() error {
	nss, err := h.db.GetAllNamespaces()
	if err != nil {
		return err
	}

	if len(nss) > 0 {
		return nil
	}

	ns, err := h.k8s.CreateNamespace(context.Background(), "wardu", nil)
	if err != nil {
		return err
	}

	return h.db.CreateNamespace(ns)
}
