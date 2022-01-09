package aws

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"time"
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

	client := s3.New(awsSession)

	_, err = client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(cfg.Bucket),
	})
	if err != nil {
		var awsErr awserr.Error
		if errors.As(err, &awsErr) && awsErr.Code() == "BucketAlreadyOwnedByYou" {
			// do nothing, pre-setup
		} else {
			return nil, err
		}
	}

	policy := `{ 
		"Version":"2012-10-17",
		"Statement":[
		   { 
			  "Action":["s3:GetObject"],
			  "Effect":"Allow",
			  "Principal":{"AWS":["*"]},
			  "Resource":["arn:aws:s3:::` + cfg.Bucket + `/*"],
			  "Sid":""
		   }
		]
	}`
	_, err = client.PutBucketPolicy(&s3.PutBucketPolicyInput{
		Bucket: aws.String(cfg.Bucket),
		Policy: aws.String(policy),
	})
	if err != nil {
		return nil, fmt.Errorf("apply bucket policy: %w", err)
	}

	s := &S3{
		bucket: cfg.Bucket,
		client: client,
	}

	return s, nil
}

func (s *S3) Put(data io.ReadSeeker, objectName, contentType string, userID string) error {
	_, err := s.client.PutObject(&s3.PutObjectInput{
		Body:        data,
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(objectName),
		ContentType: aws.String(contentType),
		Metadata: map[string]*string{
			"user-id": aws.String(userID),
		},
	})
	return err
}

func (s *S3) GetLink(objectName string) (string, error) {
	params := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectName),
	}

	req, _ := s.client.GetObjectRequest(params)

	url, err := req.Presign(15 * time.Minute) // Set link expiration time
	if err != nil {
		return "", fmt.Errorf("get link: %w", err)
	}

	return url, nil
}
