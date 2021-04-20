package aws_service

import (
	"context"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3Client struct {
	Svc *s3.S3
}

func (s3 *S3Client) PutObjectWithContext(ctx context.Context, input *s3.PutObjectInput) error {
	_, err := s3.Svc.PutObjectWithContext(ctx, input)
	return err
}

func (s3 *S3Client) PutBucketPolicy(input *s3.PutBucketPolicyInput) error {
	_, err := s3.Svc.PutBucketPolicy(input)
	return err
}
