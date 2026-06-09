package database

import "github.com/aliamerj/wardu/shared/models"

func (s *service) GetWorkerByImage(image string) (*models.Worker, error) {
	var worker models.Worker

	if err := s.db.
		Model(&models.Worker{}).
		Where("image = ?", image).
		First(&worker).Error; err != nil {
		return nil, err
	}
	return &worker, nil
}

func (s *service) CreateWorker(worker *models.Worker) error {
	return s.db.Model(&models.Worker{}).Create(worker).Error
}

func (s *service) UpdateWorker(worker *models.Worker) error {
	if err := s.db.
		Model(&models.Worker{}).
		Where("id = ?", worker.ID).
		Updates(worker).Error; err != nil {
		return err
	}
	return nil
}
