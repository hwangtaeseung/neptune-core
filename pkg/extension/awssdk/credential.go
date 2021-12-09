package awssdk

import (
	"log"
	"github.com/hwangtaeseung/neptune-core/pkg/common"
)

type AWSCredential struct {
	AwsAccessKeyId     string `json:"aws_access_key_id"`
	AwsSecretAccessKey string `json:"aws_secret_access_key"`
}

func GetAwsCredential(s3Url *S3Url) (*AWSCredential, error) {

	downloadedFile, _, err := DownloadS3(s3Url, ".", nil, nil)
	if err != nil {
		log.Printf("credential file downloading error : %v\n", err)
		return nil, err
	}

	var credential AWSCredential
	err = common.JsonFileToObject(downloadedFile, &credential)
	if err != nil {
		log.Printf("credential file parsong error : %v\n", err)
		return nil, err
	}

	return &credential, nil
}