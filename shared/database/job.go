package database

import (
	"github.com/aliamerj/wardu/shared/models"
)

func (s *service) CreateJob(job *models.Job) error {
	return s.db.Model(&models.Job{}).Create(job).Error
}

func (s *service) GetJobByID(jobId string) (*models.Job, error) {
	var job models.Job

	if err := s.db.
		Model(&models.Job{}).
		Where("id = ?", job.ID).
		First(&job).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (s *service) GetJobForExecution(
	id string,
) (*models.Job, error) {
	var job models.Job

	if err := s.db.
		Preload("Worker").
		Preload("Worker.Namespace").
		First(&job, "id = ?", id).
		Error; err != nil {
		return nil, err
	}

	return &job, nil
}
