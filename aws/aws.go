package aws

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Stanly1995/golibs/params_validator"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/labstack/gommon/log"
	"github.com/vincent-petithory/dataurl"
	"regexp"
	"time"
)

type AWSInfo struct {
	Bucket string
	URL    string
	Region string
	ID     string
	Secret string
	Token  string
}

// struct for work with AWSInfo service
type AWSConnector struct {
	timeout   time.Duration
	AWSInfo   AWSInfo
	svc       s3Client
	generator dataGenerate
	ctx       context.Context
}

// NewAWSConnector is constructor, receives aws session, bucket and aws url,
func NewAWSConnector(awsInfo AWSInfo, timeout time.Duration, ctx context.Context, svc s3Client, g dataGenerate) (*AWSConnector, error) {
	params_validator.ValidateParamsWithPanic(svc, g)
	if awsInfo.Bucket == "" {
		return nil, errors.New("bucket is empty")
	}
	if awsInfo.URL == "" {
		return nil, errors.New("aws url is empty")
	}
	if timeout <= 5*time.Second {
		return nil, errors.New("timeout is negative")
	}
	return &AWSConnector{
		timeout:   timeout,
		AWSInfo:   awsInfo,
		svc:       svc,
		generator: g,
		ctx:       ctx,
	}, nil
}

type File struct {
	content  string
	fileName string
}

func NewFile(fileObj *string) (*File, error) {
	reg := regexp.MustCompile(`^name:{(.+)},dataUrl:{(.+)}$`)

	fileParams := reg.FindStringSubmatch(*fileObj)
	if len(fileParams) != 3 {
		return nil, errors.New("invalid file object")
	}
	return &File{
		fileName: fileParams[1],
		content:  fileParams[2],
	}, nil
}

// PutFile puts input file to aws and return url for to download this file
// returns error if PutObject returns error
// Where is name - filename with extension, dataUrl - file body in dataURL format
func (awsConn *AWSConnector) PutFile(fileObj *string) (string, error) {
	file, err := NewFile(fileObj)
	if err != nil {
		return "", err
	}

	dataURLDec, err := dataurl.DecodeString(file.content)
	if err != nil {
		return "", err
	}

	var cancelFn func()

	awsConn.ctx, cancelFn = context.WithTimeout(awsConn.ctx, awsConn.timeout)

	defer cancelFn()

	select {
	case <-time.After(10 * time.Millisecond):
		log.Info("file upload started after waiting ten milliseconds")
	case <-awsConn.ctx.Done():
		log.Error(awsConn.ctx.Err()) // prints "context deadline exceeded"
	}

	uniqueFileName := fmt.Sprintf("%s_%s_%s", awsConn.generator.GenerateTime(), awsConn.generator.GenerateUUID(), file.fileName)

	err = awsConn.svc.PutObjectWithContext(awsConn.ctx, &s3.PutObjectInput{
		Body:   bytes.NewReader(dataURLDec.Data),
		Bucket: &awsConn.AWSInfo.Bucket,
		Key:    &uniqueFileName,
	})

	if err != nil {
		return "", errors.New("AWS returned error, saving file failed")
	}

	return uniqueFileName, err
}

func (awsConn *AWSConnector) SetBucketReadOnlyPolicy() error {
	readOnlyAnonUserPolicy := map[string]interface{}{
		"Statement": []map[string]interface{}{
			{
				"Sid":       "AddPerm",
				"Effect":    "Allow",
				"Principal": "*",
				"Action":    []string{"s3:GetObject"},
				"Resource":  []string{fmt.Sprintf("arn:aws:s3:::%s/*", awsConn.AWSInfo.Bucket)},
			},
		},
	}
	policy, err := json.Marshal(readOnlyAnonUserPolicy)
	if err != nil {
		return err
	}
	err = awsConn.svc.PutBucketPolicy(&s3.PutBucketPolicyInput{
		Bucket: aws.String(awsConn.AWSInfo.Bucket),
		Policy: aws.String(string(policy)),
	})

	return err
}
