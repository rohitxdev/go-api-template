package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/rohitxdev/go-api-template/env"
)

var (
	ErrFileEmpty = errors.New("file is empty")
)

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
		panic("could not load default config of S3 client: " + err.Error())
	}

	cfg.Region = env.S3_DEFAULT_REGION

	return s3.NewFromConfig(cfg)
}

var s3Client = getS3Client()

/*----------------------------------- Upload File To Bucket ----------------------------------- */

func UploadFileToBucket(ctx context.Context, bucketName string, fileName string, fileContent []byte) error {
	contentType := http.DetectContentType(fileContent)
	_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &bucketName,
		Key:         &fileName,
		Body:        bytes.NewReader(fileContent),
		ContentType: &contentType})
	return err
}

/*----------------------------------- Get File From Bucket ----------------------------------- */

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

/*----------------------------------- Delete File From Bucket ----------------------------------- */

func DeleteFileFromBucket(ctx context.Context, bucketName string, fileName string) error {
	_, err := s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: &bucketName, Key: &fileName})
	return err
}

/*----------------------------------- Get List Of Files ----------------------------------- */

type FileMetaData struct {
	FileName     string    `json:"file_name"`
	SizeInBytes  uint64    `json:"size_in_bytes"`
	LastModified time.Time `json:"last_modified"`
}

func GetFileList(ctx context.Context, bucketName string, subDir string) ([]FileMetaData, error) {
	var continuationToken *string
	var fileList []FileMetaData
	for {
		objects, err := s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: &bucketName, ContinuationToken: continuationToken, Prefix: aws.String("")})
		if err != nil {
			return nil, err
		}
		for _, file := range objects.Contents {
			fileList = append(fileList, FileMetaData{FileName: *file.Key, LastModified: *aws.Time(*file.LastModified), SizeInBytes: uint64(*file.Size)})
		}
		if !*objects.IsTruncated {
			break
		}
		continuationToken = objects.NextContinuationToken
	}
	return fileList, nil
}
