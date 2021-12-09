package awssdk

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
)

func GetKeysInBucket(region string, bucket string) ([]string, error) {

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))

	var fileList []string
	svc := s3.New(sess)
	err := svc.ListObjectsPages(&s3.ListObjectsInput{
		Bucket: aws.String(bucket),
	}, func(page *s3.ListObjectsOutput, lastPage bool) bool {
		for _, object := range page.Contents {
			fileList = append(fileList, *object.Key)
		}
		return lastPage
	})
	if err != nil {
		log.Printf("err : %v\n", err)
		return nil, err
	}
	return fileList, nil
}

func GetKeysInBucketPrefix(region, bucket, prefix string) ([]string, error) {

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))

	var fileList []string
	svc := s3.New(sess)
	err := svc.ListObjectsPages(&s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	}, func(page *s3.ListObjectsOutput, lastPage bool) bool {
		for _, object := range page.Contents {
			fileList = append(fileList, *object.Key)
		}
		return lastPage
	})
	if err != nil {
		log.Printf("err : %v\n", err)
		return nil, err
	}
	return fileList, nil
}
