package server

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/aliamerj/wardu/shared/database"
	"github.com/aliamerj/wardu/shared/env"
	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port int

	db database.Service
}

var (
	httpPort = env.GetString("HTTP_PORT", "8081")
)

func NewServer() *http.Server {
	port, _ := strconv.Atoi(httpPort)
	NewServer := &Server{
		port: port,

		db: database.New(),
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
