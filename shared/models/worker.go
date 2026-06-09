package models

import (
	"time"
)

type Worker struct {
	ID                 string `gorm:"primaryKey;type:uuid"`
	Image              string `gorm:"type:text;not null;uniqueIndex"`
	NamespaceID        string `gorm:"type:uuid;not null;index"`
	K8sDeploymentName  string `gorm:"type:text;not null;uniqueIndex"`
	MaxReplicas        int    `gorm:"not null;default:5"`
	IdleTimeoutSeconds int    `gorm:"not null;default:90"`
	LastActiveAt       *time.Time

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
