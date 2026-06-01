package server

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/aliamerj/wardu/services/api-gateway/clients"
	"github.com/aliamerj/wardu/shared/database"
	"github.com/aliamerj/wardu/shared/env"
	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port int

	db  database.Service
	srv *clients.Services
}

func NewServer(srv *clients.Services) *http.Server {
	httpPort := env.GetString("GATEWAY_HTTP_PORT", env.GetString("HTTP_PORT", "8080"))
	port, _ := strconv.Atoi(httpPort)

	NewServer := &Server{
		port: port,
		srv:  srv,
		db:   database.New(),
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
