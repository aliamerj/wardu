package handlers

import (
	"crypto/rand"
	"net/http"
	"time"

	"github.com/aliamerj/wardu/services/api-gateway/clients"
	"github.com/aliamerj/wardu/services/api-gateway/types"
	"github.com/aliamerj/wardu/shared/database"
	"github.com/labstack/echo/v5"
	"github.com/oklog/ulid/v2"
)

type Handler struct {
	db  database.Service
	srv *clients.Services
}

func New(db database.Service, srv *clients.Services) Handler {
	return Handler{
		db:  db,
		srv: srv,
	}
}

func withErr(c *echo.Context, status int, customMessage ...string) error {
	msg := getDefaultErrorMessages(status)

	if len(customMessage) > 0 && customMessage[0] != "" {
		msg = customMessage[0]
	}

	return c.JSON(status, types.ErrorResponse{
		Message: &msg,
	})
}

func getDefaultErrorMessages(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "invalid request body"

	default:
		return "Unknown error"
	}
}

func newJobID() string {
	return ulid.MustNew(
		ulid.Timestamp(time.Now()),
		ulid.Monotonic(rand.Reader, 0),
	).String()
}
