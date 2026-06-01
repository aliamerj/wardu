package clients

import "fmt"

type Services struct {
	Scheduler *schedulerClient
}

func NewServices() (*Services, error) {
	sheduler, err := newScheduler()
	if err != nil {
		return nil, err
	}

	return &Services{
		Scheduler: sheduler,
	}, nil
}

func (s *Services) CloseAll() error {
	if s.Scheduler != nil {
		if err := s.Scheduler.close(); err != nil {
			return fmt.Errorf("failed to close scheduler: %w", err)
		}
	}
	return nil
}
