package awssdk

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
	"log"
	"github.com/hwangtaeseung/neptune-core/pkg/common"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

func UploadS3(bucketName string, s3Folder string, localFileName string,
	progressCallback func(int64, int64, float32),
	endCallback func(*s3manager.UploadOutput)) (*s3manager.UploadOutput, error) {

	// open file
	file, err := os.Open(localFileName)
	if err != nil {
		log.Printf("file does not exist. (err:%v)", err)
		return nil, err
	}
	defer func() {
		if file != nil {
			if err := file.Close(); err != nil {
				log.Printf("file close err : %v", err)
			}
		}
	}()

	// get file information
	fileInfo, err := file.Stat()
	if err != nil {
		log.Printf("can not get file info. (err:%v)", err)
		return nil, err
	}

	// create session
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("ap-northeast-2"),
	}))

	_, uploadFileName := filepath.Split(localFileName)

	// upload manager
	uploader := s3manager.NewUploader(sess)
	uploader.PartSize = 40 * 1024 * 1024

	var output *s3manager.UploadOutput
	for count := 0; count < common.DefaultMaxRetryCount; count++ {

		// upload file
		output, err = uploader.Upload(
			&s3manager.UploadInput{
				Body: &progressReader{
					read:     0,
					reader:   file,
					size:     fileInfo.Size(),
					callback: progressCallback,
					time:     time.Now().Unix(),
				},
				Bucket: aws.String(bucketName),
				Key:    aws.String(fmt.Sprintf("%v/%v", s3Folder, uploadFileName)),
			},
			func(uploader *s3manager.Uploader) {
				log.Printf("upload : %v\n", uploader)
			})

		if err != nil {
			log.Printf("upload failed. (retry times:%v, err:%v)\n", count, err)
			time.Sleep(1 * time.Second)
		} else {
			// call end callback
			if endCallback != nil {
				endCallback(output)
			}
			log.Printf("upload inforrmation : %v\n", output)
			return output, nil
		}
	}

	log.Printf("upload failed finally (err:%v)\n", err)
	return nil, err
}

type progressReader struct {
	read     int64
	reader   io.Reader
	size     int64
	callback func(totalSize int64, readSize int64, percentage float32)
	time     int64
}

func (pr *progressReader) Read(p []byte) (int, error) {
	atomic.AddInt64(&pr.read, int64(len(p)))
	percentage := float32(pr.read*100 / pr.size)

	currentTime := time.Now().Unix()
	if atomic.LoadInt64(&pr.time) < currentTime {
		if pr.callback != nil {
			pr.callback(pr.size, pr.read, percentage)
		}
		atomic.StoreInt64(&pr.time, currentTime + common.DefaultProgressUpdateInterval)
	}
	return pr.reader.Read(p)
}
