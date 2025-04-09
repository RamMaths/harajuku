package awsS3

import (
	"bytes"
	"context"
	"harajuku/backend/internal/core/port"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

type AwsS3 struct {
	client s3iface.S3API
	bucket string
}

func NewAwsS3(sess *session.Session, bucket string) port.FileRepository {
	return &AwsS3 {
		client: s3.New(sess),
		bucket: bucket,
	}
}

func (a *AwsS3) Save(ctx context.Context, data []byte, name string) (string, error) {
	input := &s3.PutObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(name),
		Body:   bytes.NewReader(data),
	}

	_, err := a.client.PutObjectWithContext(ctx, input)
	if err != nil {
		return "", err
	}

	return name, nil
}

func (a *AwsS3) Get(ctx context.Context, path string) ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(path),
	}

	result, err := a.client.GetObjectWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	return io.ReadAll(result.Body)
}

func (a *AwsS3) Delete(ctx context.Context, path string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(path),
	}

	_, err := a.client.DeleteObjectWithContext(ctx, input)
	return err
}
