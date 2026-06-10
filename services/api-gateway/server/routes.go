package server

import (
	"net/http"

	"github.com/aliamerj/wardu/shared/logger"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Use(middleware.RequestID())
	e.Use(logger.RequestLogger(logger.Logger()))
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://*", "http://*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", echo.HeaderXRequestID},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	api := e.Group("api/v1")
	{
		s.CreateApiV1Routes(api)
	}

	return e
}
