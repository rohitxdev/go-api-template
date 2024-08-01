package service_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/rohitxdev/go-api-template/config"
	"github.com/rohitxdev/go-api-template/service"
)

func TestStorageService(t *testing.T) {
	ctx := context.Background()
	testFile, err := os.CreateTemp("", "test.txt")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(testFile.Name())
	defer testFile.Close()
	if err := os.WriteFile(testFile.Name(), []byte("lorem ipsum dorem"), 0666); err != nil {
		t.Error(err)
	}
	testFileContent, err := io.ReadAll(testFile)
	if err != nil {
		t.Error(err)
	}

	t.Run("Upload file to bucket", func(t *testing.T) {
		err := service.UploadFileToBucket(ctx, config.S3_BUCKET_NAME, testFile.Name(), testFileContent)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("Get file from bucket", func(t *testing.T) {
		fileContent, err := service.GetFileFromBucket(ctx, config.S3_BUCKET_NAME, testFile.Name())
		if err != nil {
			t.Error(err)
		}
		if !bytes.Equal(fileContent, testFileContent) {
			t.Error(service.ErrFileEmpty)
		}
	})

	t.Run("Delete file from bucket", func(t *testing.T) {
		err := service.DeleteFileFromBucket(ctx, config.S3_BUCKET_NAME, testFile.Name())
		if err != nil {
			t.Error(err)
		}
		_, err = service.GetFileFromBucket(ctx, config.S3_BUCKET_NAME, testFile.Name())
		if err == nil {
			t.Error(err)
		}
	})
}
