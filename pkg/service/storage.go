package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	ErrFileEmpty = errors.New("file is empty")
)

func InitS3Client(endpoint string, region string, accessKeyId string, accessKeySecret string) (*s3.Client, error) {
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: endpoint,
		}, nil
	})

	cfg, err := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithEndpointResolverWithOptions(r2Resolver),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret, "")),
	)
	if err != nil {
		return nil, errors.Join(errors.New("could not load default config of S3 client"), err)
	}

	cfg.Region = region

	return s3.NewFromConfig(cfg), nil
}

type FileStorage struct {
	client *s3.Client
}

func NewFileStorage(endpoint string, region string, accessKeyId string, accessKeySecret string) (*FileStorage, error) {
	s := new(FileStorage)
	client, err := InitS3Client(endpoint, region, accessKeyId, accessKeySecret)
	if err != nil {
		return nil, err
	}
	s.client = client
	return s, nil
}

/*----------------------------------- Upload File To Bucket ----------------------------------- */

func (s *FileStorage) Upload(ctx context.Context, bucketName string, fileName string, fileContent []byte) error {
	contentType := http.DetectContentType(fileContent)
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &bucketName,
		Key:         &fileName,
		Body:        bytes.NewReader(fileContent),
		ContentType: &contentType})
	return err
}

/*----------------------------------- Get File From Bucket ----------------------------------- */

func (s *FileStorage) Get(ctx context.Context, bucketName string, fileName string) ([]byte, error) {
	obj, err := s.client.GetObject(ctx, &s3.GetObjectInput{Bucket: &bucketName, Key: &fileName})
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

func (s *FileStorage) Delete(ctx context.Context, bucketName string, fileName string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: &bucketName, Key: &fileName})
	return err
}

/*----------------------------------- Get List Of Files ----------------------------------- */

type FileMetaData struct {
	LastModified time.Time `json:"last_modified"`
	FileName     string    `json:"file_name"`
	SizeInBytes  uint64    `json:"size_in_bytes"`
}

func (s *FileStorage) GetList(ctx context.Context, bucketName string, subDir string) ([]FileMetaData, error) {
	var continuationToken *string
	var fileList []FileMetaData
	for {
		objects, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: &bucketName, ContinuationToken: continuationToken, Prefix: aws.String(subDir)})
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
