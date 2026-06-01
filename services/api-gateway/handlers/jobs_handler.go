package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/aliamerj/wardu/services/api-gateway/clients"
	"github.com/aliamerj/wardu/services/api-gateway/types"
	"github.com/labstack/echo/v5"
)

func (h *Handler) AddJobsEndPoints(e *echo.Group) {
	e.POST("", h.submitJob)
}

func (h *Handler) submitJob(c *echo.Context) error {
	ctx := c.Request().Context()
	var req types.SubmitJobRequest

	if err := c.Bind(&req); err != nil {
		return withErr(c, http.StatusBadRequest)
	}

	if strings.TrimSpace(req.Worker) == "" {
		return withErr(c, http.StatusBadRequest)
	}

	if len(req.Payload) == 0 {
		return withErr(c, http.StatusBadRequest)
	}

	payload, err := json.Marshal(req.Payload)
	if err != nil {
		return withErr(c, http.StatusBadRequest)
	}

	job := clients.Job{
		JobId:   newJobID(),
		Payload: payload,
		Worker:  req.Worker,
	}
	if req.Priority != nil {
		job.Priority = int64(*req.Priority)
	} else {
		job.Priority = 1
	}

	res, err := h.srv.Scheduler.CreateJob(ctx, &job)
	if err != nil {
		return withErr(c, http.StatusServiceUnavailable, "scheduler unavailable")
	}

	return c.JSON(http.StatusAccepted, types.SubmitJobResponse{
		JobId:  res.JobId,
		Status: types.JobStatusPending,
	})
}
