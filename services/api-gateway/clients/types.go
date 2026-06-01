package clients

import (
	pb "github.com/aliamerj/wardu/shared/proto/scheduler"
)

type Job struct {
	JobId    string
	Payload  []byte
	Priority int64
	Worker   string
}

func (j *Job) toProto() *pb.CreateJobRequest {
	return &pb.CreateJobRequest{
		JobId:    j.JobId,
		Payload:  j.Payload,
		Worker:   j.Worker,
		Priority: j.Priority,
	}
}
