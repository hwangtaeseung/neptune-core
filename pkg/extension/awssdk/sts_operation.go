package awssdk

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"log"
	"github.com/hwangtaeseung/neptune-core/pkg/common"
	"strings"
	"time"
)

func GetS3CredentialQueryString(region, serviceName string) (string, error) {

	if headers, err := GetSignature(region, serviceName); err != nil {
		log.Printf("get signature error (err:%v)\n", err)
		return "", err
	} else {
		contentSha256 := hex.EncodeToString(GetSha256("", []byte(headers.SecretAccessKey)))
		log.Printf("* empty string sha256 : %v\n", contentSha256)
		return common.QueryString(map[string]string{
			"authorization": fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%v/%v/%v/%v/aws4_request, SignedHeaders=host;x-amz-date;x-amz-security-token;x-amz-content-sha256, Signature=%v",
				headers.AccessKeyId, headers.DateStamp, region, serviceName, headers.Signature),
			"x-amz-security-token": headers.SessionToken,
			"x-amz-content-sha256": contentSha256,
			"x-amz-date" : strings.ReplaceAll(strings.ReplaceAll(time.Now().Format(time.RFC3339), "-", ""), ":", "") ,
		}), nil
	}
}

func GetSignature(region, serviceName string) (*CredentialHeaders, error) {

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))

	svc := sts.New(sess)
	token, err := svc.GetSessionToken(&sts.GetSessionTokenInput{
		DurationSeconds: aws.Int64(int64(2 * 60 * 60)),
	})
	if err != nil {
		log.Printf("get federation totken error. (err:%v)", err)
		return nil, err
	}

	dateStamp := getDateStamp()
	return &CredentialHeaders{
		AccessKeyId:  *token.Credentials.AccessKeyId,
		Signature:    getSignature(*token.Credentials.SecretAccessKey, dateStamp, region, serviceName),
		SessionToken: *token.Credentials.SessionToken,
		SecretAccessKey: *token.Credentials.SecretAccessKey,
		DateStamp:    dateStamp,
	}, nil
}

type CredentialHeaders struct {
	SessionToken    string
	AccessKeyId     string
	SecretAccessKey string
	Signature       string
	DateStamp       string
}

func getSignature(key, dateStamp, region, name string) string {
	kSecret := []byte(fmt.Sprintf("AWS4%v", key))
	kDate := GetSha256(dateStamp, kSecret)
	kRegion := GetSha256(region, kDate)
	kService := GetSha256(name, kRegion)
	return hex.EncodeToString(GetSha256("aws4_request", kService))
}

func GetSha256(data string, secret []byte) []byte {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(data))
	return h.Sum(nil)
}

func getDateStamp() string {
	today := time.Now().UTC()
	return fmt.Sprintf("%04d%02d%02d", today.Year(), today.Month(), today.Day())
}
