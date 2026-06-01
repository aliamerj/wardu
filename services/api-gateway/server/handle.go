package server

import (
	"net/http"

	"github.com/aliamerj/wardu/services/api-gateway/handlers"
	"github.com/labstack/echo/v5"
)

func (s *Server) CreateApiV1Routes(e *echo.Group) {
	hand := handlers.New(s.db, s.srv)
	e.GET("/health", s.healthHandler)
	jobs := e.Group("/jobs")
	{
		hand.AddJobsEndPoints(jobs)
	}
}

func (s *Server) healthHandler(c *echo.Context) error {
	// TODO: add more hearth check
	return c.JSON(http.StatusOK, s.db.Health())
}
