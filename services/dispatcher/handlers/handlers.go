package handlers

import (
	"context"

	"github.com/aliamerj/wardu/shared/database"
	"github.com/aliamerj/wardu/shared/k8s"
	r "github.com/aliamerj/wardu/shared/rabbitmq"
)

func ExecuteJob(
	ctx context.Context,
	db database.Service,
	k8s *k8s.Client,
	jm r.JobMessage,
) error {
	// TODO : Load Job -> Load Worker -> Scale Deployment -> Wait Ready if not running -> run -> Save Result -> Update Worker Activity -> Success
	return nil
}
