package models

import (
	"encoding/json"

	"github.com/aliamerj/wardu/services/api-gateway/types"
	pb "github.com/aliamerj/wardu/shared/proto/scheduler"
)

type Job struct {
	ID                 string   `gorm:"primaryKey;type:varchar(255)"`
	Payload            []byte   `gorm:"type:bytea"`
	Priority           int64    `gorm:"not null;default:0"`
	Autorun            bool     `gorm:"not null;default:false"`
	Entrypoint         []string `gorm:"serializer:json"`
	IdleTimeoutSeconds float32
	Image              string `gorm:"not null"`
	MaxAttempts        float32
	Namespace          string `gorm:"index"`
	TimeoutSeconds     float32
}

func BuildNewJob(jobReq types.SubmitJobRequest) (*Job, error) {
	payload, err := json.Marshal(jobReq.Payload)
	if err != nil {
		return nil, err
	}

	job := Job{
		Image:   jobReq.Image,
		Payload: payload,
	}

	if jobReq.Priority != nil {
		job.Priority = int64(*jobReq.Priority)
	} else {
		job.Priority = 1
	}

	if jobReq.Autorun != nil {
		job.Autorun = *jobReq.Autorun
	}

	if jobReq.Entrypoint != nil {
		job.Entrypoint = *jobReq.Entrypoint
	}

	if jobReq.Namespace != nil {
		job.Namespace = *jobReq.Namespace
	} else {
		job.Namespace = "wardu"
	}

	if jobReq.IdleTimeoutSeconds != nil {
		job.IdleTimeoutSeconds = *jobReq.IdleTimeoutSeconds
	}

	if jobReq.MaxAttempts != nil {
		job.MaxAttempts = *jobReq.MaxAttempts
	}

	if jobReq.TimeoutSeconds != nil {
		job.TimeoutSeconds = *jobReq.TimeoutSeconds
	}

	return &job, nil
}

func (j *Job) ToProto() *pb.CreateJobRequest {
	return &pb.CreateJobRequest{
		Image:              j.Image,
		Payload:            j.Payload,
		Priority:           j.Priority,
		Autorun:            j.Autorun,
		Entrypoint:         j.Entrypoint,
		IdleTimeoutSeconds: j.IdleTimeoutSeconds,
		MaxAttempts:        j.MaxAttempts,
		Namespace:          j.Namespace,
		TimeoutSeconds:     j.TimeoutSeconds,
	}
}
