package aws_service

import (
	"context"
	"github.com/aws/aws-sdk-go/service/s3"
)

//go:generate mockgen -source=interface.go -destination=mocks_test.go -package=aws
type s3Client interface {
	PutObjectWithContext(ctx context.Context, input *s3.PutObjectInput) error
	PutBucketPolicy(input *s3.PutBucketPolicyInput) error
}

type dataGenerate interface {
	GenerateTime() string
	GenerateUUID() string
}
