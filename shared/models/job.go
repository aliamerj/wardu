package models

import (
	"encoding/json"
	"time"

	"github.com/aliamerj/wardu/services/api-gateway/types"
	pb "github.com/aliamerj/wardu/shared/proto/scheduler"
)

type status string

const (
	StatusPending   status = "pending"
	StatusReady     status = "ready"
	StatusRunning   status = "running"
	StatusSucceeded status = "succeeded"
	StatusFailed    status = "failed"
	StatusDead      status = "dead"
	StatusCancelled status = "cancelled"
)

type Job struct {
	ID         string   `gorm:"primaryKey;type:varchar(255)"`
	WorkerID   string   `gorm:"type:uuid;not null;index"`
	Namespace  string   `gorm:"type:uuid;not null;index"`
	Status     status   `gorm:"type:varchar(255);not null;default:pending"`
	Autorun    bool     `gorm:"not null;default:false"`
	Entrypoint []string `gorm:"serializer:json"`
	Payload    []byte   `gorm:"type:bytea"`

	IdleTimeoutSeconds float32
	MaxAttempts        float32
	TimeoutSeconds     float32

	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func BuildJobProto(jobReq types.SubmitJobRequest) (*pb.CreateJobRequest, error) {
	payload, err := json.Marshal(jobReq.Payload)
	if err != nil {
		return nil, err
	}

	job := pb.CreateJobRequest{
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

func BuildJobFromProto(proto *pb.CreateJobRequest) *Job {
	return &Job{
		Payload:            proto.GetPayload(),
		Autorun:            proto.GetAutorun(),
		Entrypoint:         proto.GetEntrypoint(),
		IdleTimeoutSeconds: proto.GetIdleTimeoutSeconds(),
		MaxAttempts:        proto.GetMaxAttempts(),
		Namespace:          proto.GetNamespace(),
		TimeoutSeconds:     proto.GetTimeoutSeconds(),
	}
}
