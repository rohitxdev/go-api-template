package storage

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	ErrFileEmpty = errors.New("file is empty")
)

type Client struct {
	client    *s3.Client
	presigner *s3.PresignClient
}

func New(endpoint string, region string, accessKeyId string, accessKeySecret string) (*Client, error) {
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: endpoint,
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret, "")),
	)
	if err != nil {
		return nil, errors.Join(errors.New("could not load default config of S3 client"), err)
	}

	cfg.Region = region

	s3Client := s3.NewFromConfig(cfg)

	client := Client{
		client:    s3Client,
		presigner: s3.NewPresignClient(s3Client),
	}

	return &client, nil
}

/*----------------------------------- Upload File To Bucket ----------------------------------- */

func (s *Client) PresignPutObject(ctx context.Context, bucketName string, fileName string, contentType string) (*v4.PresignedHTTPRequest, error) {
	request, err := s.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      &bucketName,
		Key:         &fileName,
		ContentType: &contentType})
	return request, err
}

/*----------------------------------- Get File From Bucket ----------------------------------- */

func (s *Client) PresignGetObject(ctx context.Context, bucketName string, fileName string) (*v4.PresignedHTTPRequest, error) {
	return s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{Bucket: &bucketName, Key: &fileName}, func(po *s3.PresignOptions) { po.Expires = time.Minute * 2 })
}

/*----------------------------------- Delete File From Bucket ----------------------------------- */

func (s *Client) PresignDeleteObject(ctx context.Context, bucketName string, fileName string) (*v4.PresignedHTTPRequest, error) {
	return s.presigner.PresignDeleteObject(ctx, &s3.DeleteObjectInput{Bucket: &bucketName, Key: &fileName})
}

/*----------------------------------- Get List Of Files ----------------------------------- */

type FileMetaData struct {
	LastModified time.Time `json:"last_modified"`
	FileName     string    `json:"file_name"`
	SizeInBytes  uint64    `json:"size_in_bytes"`
}

func (s *Client) GetList(ctx context.Context, bucketName string, subDir string) ([]FileMetaData, error) {
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
