package awssdk

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/batch"
	"log"
	"neptune-core/pkg/common"
)

func SubmitJob(region, jobDefinition, jobName, jobQueue string, envs []*batch.KeyValuePair,
	arrayProperties *batch.ArrayProperties, dependOn []*batch.JobDependency) (*batch.SubmitJobOutput, error) {

	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(region),
	}))

	input := &batch.SubmitJobInput{
		DependsOn:     dependOn,
		JobDefinition: aws.String(jobDefinition),
		JobName:       aws.String(jobName),
		JobQueue:      aws.String(jobQueue),
		ContainerOverrides: &batch.ContainerOverrides{
			Environment: envs,
		},
		ArrayProperties: arrayProperties,
	}

	service := batch.New(sess)
	if output, err := service.SubmitJob(input); err != nil {
		log.Printf("job submit error : %v", err)
		return nil, err
	} else {
		log.Printf("job submit ok : %v", common.ToJsonAsString(output))
		return output, nil
	}
}

func TerminateJob(region, jobId, reason string) (*batch.TerminateJobOutput, error) {

	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(region),
	}))

	input := &batch.TerminateJobInput{
		JobId:     aws.String(jobId),
		Reason:    aws.String(reason),
	}

	service := batch.New(sess)
	if output, err := service.TerminateJob(input); err != nil {
		log.Printf("terminate job error : %v", err)
		return nil, err
	} else {
		log.Printf("terminate job ok : %v", common.ToJsonAsString(output))
		return output, nil
	}
}
