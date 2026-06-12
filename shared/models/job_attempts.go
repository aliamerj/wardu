package models

import "time"

type AttemptStatus string

const (
	AttemptStatusRunning   AttemptStatus = "running"
	AttemptStatusSucceeded AttemptStatus = "succeeded"
	AttemptStatusFailed    AttemptStatus = "failed"
	AttemptStatusDead      AttemptStatus = "dead"
	AttemptStatusCancelled AttemptStatus = "cancelled"
)

type JobAttempt struct {
	ID         string        `gorm:"primaryKey;type:varchar(255)"`
	JobID      string        `gorm:"type:uuid;not null;index"`
	Attempt    int           `gorm:"not null"`
	Status     AttemptStatus `gorm:"tupe:text;not null"`
	WorkerPod  string        `gorm:"type:text"`
	Result     []byte        `gorm:"type:bytea"`
	Error      string        `gorm:"type:text"`
	StartedAt  time.Time     `gorm:"autoCreateTime"`
	FinishedAt time.Time
}
