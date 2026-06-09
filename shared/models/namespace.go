package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Namespace struct {
	ID   string `gorm:"primaryKey"`
	Name string `gorm:"type:text;not null;unique"`
	DNS  string `gorm:"type:text;not null;unique"`

	MaxWorkers        int `gorm:"not null;default:10"`
	MaxConcurrentJobs int `gorm:"not null;default:50"`
	MaxPods           int `gorm:"not null;default:20"`

	CPURequestMilli int `gorm:"not null;default:100"`
	CPULimitMilli   int `gorm:"not null;default:500"`

	MemoryRequestMB int `gorm:"not null;default:256"`
	MemoryLimitMB   int `gorm:"not null;default:512"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (ns *Namespace) BeforeCreate(tx *gorm.DB) error {
	if ns.ID == "" {
		ns.ID = uuid.NewString()
	}

	return nil
}
