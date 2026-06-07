package server

import (
	"context"

	pb "github.com/aliamerj/wardu/shared/proto/scheduler"
)

type JobModel struct {
	JobId string `bson:"jobID"`
}

type SchedulerService interface {
	CreateJob(ctx context.Context, fare *pb.CreateJobRequest) (*JobModel, error)
}

func (t *JobModel) ToProto() *pb.CreateJobResponse {
	return &pb.CreateJobResponse{
		JobId: t.JobId,
	}
}
