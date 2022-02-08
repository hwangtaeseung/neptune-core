package awssdk

import (
	"fmt"
	"github.com/hwangtaeseung/neptune-core/pkg/common"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func DownloadS3(s3Url *S3Url, tempDir string,
	progressCallback func(int64, int64, float32), endCallback func(int64)) (string, int64, error) {
	return DownloadS3Partially(s3Url, tempDir, "", progressCallback, endCallback)
}

func DownloadS3Partially(s3Url *S3Url, tempDir string, partialRange string,
	progressCallback func(int64, int64, float32), endCallback func(int64)) (string, int64, error) {

	// make temp Dir
	_ = os.Mkdir(tempDir, os.ModePerm)

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(s3Url.Region),
	}))

	temp, err := ioutil.TempFile(tempDir, "s3-download-tmp-")
	if err != nil {
		return "", 0, err
	}
	defer func() {
		if temp == nil {
			return
		}
		if err := temp.Close(); err != nil {
			log.Printf("temporary file close error.")
		} else {
			log.Printf("temporary file close success.")
		}
	}()

	tempFileName := temp.Name()
	log.Printf("temp file name : %v\n", tempFileName)

	params := &s3.GetObjectInput{
		Bucket: aws.String(s3Url.InputBucket),
		Key:    aws.String(s3Url.Key),
		Range:  aws.String(partialRange),
	}

	downloader := s3manager.NewDownloader(sess)

	var downloadedSize int64
	for count := 0; count < common.DefaultMaxRetryCount; count++ {

		client := s3.New(sess)
		var fileSize int64
		fileSize, err = getS3FileSize(client, s3Url.InputBucket, s3Url.Key)
		if err != nil {
			log.Printf("get file size error. (err:%v, s3Url:%+v). retry %d times \n", err, s3Url, count)
			// delay...
			time.Sleep(time.Second)
			continue
		}

		wd, _ := os.Getwd()
		disk, err := common.DiskUsage(wd)
		if float64(disk.Free)*0.95 < float64(fileSize) {
			return "", 0, fmt.Errorf("out of disk space (downloadFileSize:%v, diskDize:%v)", fileSize, disk.Free)
		}

		writer := &progressWriter{
			writer:   temp,
			size:     fileSize,
			written:  0,
			callback: progressCallback,
		}

		downloadedSize, err = downloader.Download(writer, params)
		if err != nil {
			log.Printf("Download failed! retry %d times\n", count+1)

			// delay...
			time.Sleep(time.Second)

		} else {
			// downloading complete
			if endCallback != nil {
				endCallback(downloadedSize)
			}
			log.Printf("downlod return value : %v, %v", downloadedSize, tempFileName)

			// rename temp file name
			_, fileName := filepath.Split(s3Url.Key)
			fileName = fmt.Sprintf("%v/%v", tempDir, fileName)
			if err := os.Rename(temp.Name(), fileName); err != nil {
				log.Printf("rename error")
				return "", 0, err
			}
			log.Println("File downloaded! Available at:", fileName)

			return fileName, downloadedSize, nil
		}
	}

	log.Printf("Download failed finally!! (err:%v)\n", err)

	return "", 0, err
}

/////////////////////////
// progress writer class
/////////////////////////
type progressWriter struct {
	written  int64
	writer   io.WriterAt
	size     int64
	callback func(size int64, written int64, percent float32)
	time     int64
}

func (pw *progressWriter) WriteAt(p []byte, off int64) (int, error) {
	atomic.AddInt64(&pw.written, int64(len(p)))
	percentage := float32(pw.written * 100 / pw.size)

	currentTime := time.Now().Unix()
	if atomic.LoadInt64(&pw.time) < currentTime {
		if pw.callback != nil {
			pw.callback(pw.size, pw.written, percentage)
		}
		atomic.StoreInt64(&pw.time, currentTime+common.DefaultProgressUpdateInterval)
	}

	return pw.writer.WriteAt(p, off)
}

func byteCountDecimal(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

func getS3FileSize(svc *s3.S3, bucket string, key string) (int64, error) {
	params := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	resp, err := svc.HeadObject(params)
	if err != nil {
		return 0, err
	}

	return *resp.ContentLength, nil
}

func GetPreSignedUrl(s3Url *S3Url, minutes int64) (string, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(s3Url.Region),
	})
	if err != nil {
		log.Printf("create session error.\n")
		return "", err
	}

	svc := s3.New(sess)
	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s3Url.InputBucket),
		Key:    aws.String(s3Url.Key),
	})

	preSignedUrl, err := req.Presign(time.Duration(minutes) * time.Minute)
	if err != nil {
		log.Println("Failed to sign request", err)
		return "", err
	}
	return preSignedUrl, nil
}
