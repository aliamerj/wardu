package clients

import (
	"fmt"

	zlog "github.com/rs/zerolog/log"
)

type Services struct {
	Scheduler *schedulerClient
}

func NewServices() (*Services, error) {
	scheduler, err := newScheduler()
	if err != nil {
		return nil, err
	}

	zlog.Info().Msg("api gateway service clients initialized")
	return &Services{Scheduler: scheduler}, nil
}

func (s *Services) CloseAll() error {
	if s.Scheduler != nil {
		if err := s.Scheduler.close(); err != nil {
			return fmt.Errorf("failed to close scheduler: %w", err)
		}
	}
	return nil
}
