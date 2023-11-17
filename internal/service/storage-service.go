package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/rohitxdev/go-api-template/internal/env"
)

var (
	ErrFileEmpty = errors.New("file is empty")
)

var s3Client = getS3Client()

func getS3Client() *s3.Client {
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: env.S3_ENDPOINT,
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(env.AWS_ACCESS_KEY_ID, env.AWS_ACCESS_KEY_SECRET, "")),
	)
	if err != nil {
		panic(err)
	}

	cfg.Region = env.S3_DEFAULT_REGION

	return s3.NewFromConfig(cfg)
}

func UploadFileToBucket(ctx context.Context, bucketName string, fileName string, fileContent []byte) error {
	contentType := http.DetectContentType(fileContent)
	_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &bucketName,
		Key:         &fileName,
		Body:        bytes.NewReader(fileContent),
		ContentType: &contentType})
	return err
}

func GetFileFromBucket(ctx context.Context, bucketName string, fileName string) ([]byte, error) {
	obj, err := s3Client.GetObject(ctx, &s3.GetObjectInput{Bucket: &bucketName, Key: &fileName})
	if err != nil {
		return nil, err
	}

	fileContent, err := io.ReadAll(obj.Body)
	if err != nil {
		return nil, err
	}
	defer obj.Body.Close()

	if len(fileContent) == 0 {
		return nil, ErrFileEmpty
	}

	return fileContent, nil
}

func DeleteFileFromBucket(ctx context.Context, bucketName string, fileName string) error {
	_, err := s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: &bucketName, Key: &fileName})
	return err
}
