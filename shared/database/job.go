package database

import "github.com/aliamerj/wardu/shared/models"

func (s *service) CreateJob(job *models.Job) error {
	return s.db.Model(&models.Job{}).Create(job).Error
}
