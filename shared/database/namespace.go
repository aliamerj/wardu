package database

import "github.com/aliamerj/wardu/shared/models"

func (s *service) CreateNamespace(ns *models.Namespace) error {
	return s.db.Model(&models.Namespace{}).Create(ns).Error
}

func (s *service) DeleteNamespace(name string) error {
	return s.db.Delete(&models.Namespace{}, "name = ?", name).Error
}

func (s *service) GetAllNamespaces() ([]*models.Namespace, error) {
	var nss []*models.Namespace
	if err := s.db.
		Model(&models.Namespace{}).
		Find(&nss).Error; err != nil {
		return nil, err
	}

	return nss, nil
}

func (s *service) GetNamespaceByName(name string) (*models.Namespace, error) {
	var ns models.Namespace

	if err := s.db.
		Model(&models.Namespace{}).
		Where("name = ?", name).
		First(&ns).Error; err != nil {
		return nil, err
	}
	return &ns, nil
}

func (s *service) UpdateNamespace(name string, newNS models.Namespace) (*models.Namespace, error) {
	var ns models.Namespace
	if err := s.db.
		Model(&models.Namespace{}).
		Where("name = ?", name).
		Updates(newNS).
		First(&ns).Error; err != nil {
		return nil, err
	}
	return &ns, nil
}
