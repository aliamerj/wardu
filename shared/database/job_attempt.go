package database

import "github.com/aliamerj/wardu/shared/models"

func (s *service) CreateJobAttempt(jt *models.JobAttempt) error {
	return s.db.Model(models.JobAttempt{}).Create(jt).Error
}

func (s *service) UpdateJobAttempt(jt *models.JobAttempt) (*models.JobAttempt, error) {
	var jobAttempt models.JobAttempt
	if err := s.db.
		Model(&models.JobAttempt{}).
		Where("id = ?", jt.ID).
		Updates(jt).
		First(&jobAttempt).Error; err != nil {
		return nil, err
	}
	return &jobAttempt, nil
}
