package database

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aliamerj/wardu/shared/models"
	zlog "github.com/rs/zerolog/log"
)

type Service interface {
	Health() map[string]string
	CreateNamespace(ns *models.Namespace) error
	DeleteNamespace(name string) error
	GetAllNamespaces() ([]*models.Namespace, error)
	GetNamespaceByName(name string) (*models.Namespace, error)
	UpdateNamespace(name string, newNS models.Namespace) (*models.Namespace, error)
	GetWorkerByImage(image string) (*models.Worker, error)
	CreateJob(job *models.Job) error
	CreateWorker(worker *models.Worker) error
	UpdateWorker(worker *models.Worker) error
}

func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stats := make(map[string]string)
	stats["status"] = "down"

	sqlDB, err := s.db.DB()
	if err != nil {
		stats["error"] = fmt.Sprintf("db down: %v", err)
		zlog.Error().Err(err).Msg("failed to get sql.DB handle")
		return stats
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		stats["error"] = fmt.Sprintf("db down: %v", err)
		zlog.Error().Err(err).Msg("database health check failed")
		return stats
	}

	stats["status"] = "up"
	stats["message"] = "It's healthy"

	dbStats := sqlDB.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	if dbStats.OpenConnections > 40 || dbStats.WaitCount > 1000 || dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 || dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		zlog.Warn().
			Int("open_connections", dbStats.OpenConnections).
			Int("in_use", dbStats.InUse).
			Int("idle", dbStats.Idle).
			Int64("wait_count", dbStats.WaitCount).
			Dur("wait_duration", dbStats.WaitDuration).
			Int64("max_idle_closed", dbStats.MaxIdleClosed).
			Int64("max_lifetime_closed", dbStats.MaxLifetimeClosed).
			Msg("database health check indicates pressure")
	}

	if dbStats.OpenConnections > 40 {
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}
