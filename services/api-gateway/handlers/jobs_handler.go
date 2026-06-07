package handlers

import (
	"net/http"

	"github.com/aliamerj/wardu/services/api-gateway/types"
	"github.com/aliamerj/wardu/shared/models"
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

	job, err := models.BuildNewJob(req)
	if err != nil {
		return withErr(c, http.StatusBadRequest)
	}

	res, err := h.srv.Scheduler.CreateJob(ctx, job)
	if err != nil {
		return withErr(c, http.StatusServiceUnavailable, "scheduler unavailable")
	}

	return c.JSON(http.StatusAccepted, types.SubmitJobResponse{
		JobId:  res.JobId,
		Status: types.JobStatusPending,
	})
}
