package handlers

import (
	"net/http"

	"github.com/aliamerj/wardu/services/api-gateway/types"
	"github.com/aliamerj/wardu/shared/models"
	"github.com/labstack/echo/v5"
	zlog "github.com/rs/zerolog/log"
	"google.golang.org/grpc/metadata"
)

func (h *Handler) AddJobsEndPoints(e *echo.Group) {
	e.POST("", h.submitJob)
}

func (h *Handler) submitJob(c *echo.Context) error {
	ctx := c.Request().Context()
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)
	if requestID != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-request-id", requestID)
	}

	var req types.SubmitJobRequest
	if err := c.Bind(&req); err != nil {
		zlog.Warn().Str("request_id", requestID).Err(err).Msg("invalid job submission payload")
		return withErr(c, http.StatusBadRequest)
	}

	job, err := models.BuildJobProto(req)
	if err != nil {
		zlog.Warn().Str("request_id", requestID).Err(err).Msg("failed to translate job submission payload")
		return withErr(c, http.StatusBadRequest)
	}

	zlog.Info().
		Str("request_id", requestID).
		Str("image", req.Image).
		Str("namespace", job.Namespace).
		Bool("autorun", job.Autorun).
		Msg("submitting job to scheduler")

	res, err := h.srv.Scheduler.CreateJob(ctx, job)
	if err != nil {
		zlog.Error().Str("request_id", requestID).Err(err).Str("image", req.Image).Msg("scheduler rejected job submission")
		return withErr(c, http.StatusBadRequest, err.Error())
	}

	zlog.Info().
		Str("request_id", requestID).
		Str("job_id", res.JobId).
		Msg("job submission accepted")

	return c.JSON(http.StatusAccepted, types.SubmitJobResponse{
		JobId:  res.JobId,
		Status: types.JobStatusPending,
	})
}
