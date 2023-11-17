package service_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/rohitxdev/go-api-template/internal/env"
	"github.com/rohitxdev/go-api-template/internal/service"
)

func TestStorageService(t *testing.T) {
	ctx := context.Background()
	testFileAbsolutePath := env.PROJECT_ROOT + "/public/images/go-fast.png"
	testFileName := filepath.Base(testFileAbsolutePath)
	testFileContent, err := os.ReadFile(testFileAbsolutePath)
	if err != nil {
		t.Error(err)
	}

	t.Run("Upload file to bucket", func(t *testing.T) {
		err := service.UploadFileToBucket(ctx, env.S3_BUCKET_NAME, testFileName, testFileContent)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("Get file from bucket", func(t *testing.T) {
		fileContent, err := service.GetFileFromBucket(ctx, env.S3_BUCKET_NAME, testFileName)
		if err != nil {
			t.Error(err)
		}
		if !bytes.Equal(fileContent, testFileContent) {
			t.Error(service.ErrFileEmpty)
		}
	})

	t.Run("Delete file from bucket", func(t *testing.T) {
		err := service.DeleteFileFromBucket(ctx, env.S3_BUCKET_NAME, testFileName)
		if err != nil {
			t.Error(err)
		}
		_, err = service.GetFileFromBucket(ctx, env.S3_BUCKET_NAME, testFileName)
		if err == nil {
			t.Error(err)
		}
	})
}
