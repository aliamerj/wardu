package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/aliamerj/wardu/services/api-gateway/clients"
	serverpkg "github.com/aliamerj/wardu/services/api-gateway/server"
	"github.com/aliamerj/wardu/shared/logger"
	"github.com/rs/zerolog"
)

func main() {
	log := logger.Setup("api-gateway")
	log.Info().Msg("starting api gateway")

	srv, err := clients.NewServices()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize api gateway clients")
	}

	httpServer := serverpkg.NewServer(srv)
	log.Info().Str("addr", httpServer.Addr).Msg("api gateway server listening")
	done := make(chan struct{}, 1)

	go gracefulShutdown(log, httpServer, done)

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("http server stopped unexpectedly")
	}

	<-done

	if err := srv.CloseAll(); err != nil {
		log.Error().Err(err).Msg("failed to close client connections")
	}

	log.Info().Msg("api gateway exited cleanly")
}

func gracefulShutdown(log zerolog.Logger, apiServer *http.Server, done chan<- struct{}) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Info().Msg("shutdown signal received, stopping api gateway")
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("api gateway shutdown failed")
	}

	log.Info().Msg("api gateway shutdown complete")
	done <- struct{}{}
}
