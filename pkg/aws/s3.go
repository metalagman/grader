package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3 struct {
	bucket string
	client *s3.S3
}

func NewS3(cfg Config) (*S3, error) {
	awsConfig := &aws.Config{
		Credentials: credentials.NewStaticCredentials(cfg.Key, cfg.Secret, ""),
		Region:      aws.String(cfg.Region),
		Endpoint:    aws.String(cfg.Endpoint),
	}
	awsSession, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("s3 session: %w", err)
	}

	s3client := s3.New(awsSession)

	s := &S3{
		bucket: cfg.Bucket,
		client: s3client,
	}

	return s, nil
}
