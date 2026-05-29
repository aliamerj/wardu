package server

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func (s *Server) CreateRoutes(e *echo.Echo) {
	e.GET("/", s.HelloWorldHandler)
	e.GET("/health", s.healthHandler)
}

func (s *Server) HelloWorldHandler(c *echo.Context) error {
	resp := map[string]string{
		"message": "Hello World",
	}

	return c.JSON(http.StatusOK, resp)
}

func (s *Server) healthHandler(c *echo.Context) error {
	return c.JSON(http.StatusOK, s.db.Health())
}
