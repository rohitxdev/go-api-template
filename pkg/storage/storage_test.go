package storage_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/rohitxdev/go-api-template/internal/config"
	"github.com/rohitxdev/go-api-template/pkg/storage"
)

func TestStorageService(t *testing.T) {
	cfg, err := config.Load("../.env")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	fs, err := storage.New(cfg.S3Endpoint, cfg.S3DefaultRegion, cfg.AwsAccessKeyId, cfg.AwsAccessKeySecret)
	if err != nil {
		t.Fatal(err)
	}

	testFile, err := os.CreateTemp("", "test.txt")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(testFile.Name())
	defer testFile.Close()
	if err = os.WriteFile(testFile.Name(), []byte("lorem ipsum dorem"), 0666); err != nil {
		t.Error(err)
	}
	testFileContent, err := io.ReadAll(testFile)
	if err != nil {
		t.Error(err)
	}

	t.Run("Upload file to bucket", func(t *testing.T) {
		err := fs.Upload(ctx, cfg.S3Endpoint, testFile.Name(), testFileContent)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("Get file from bucket", func(t *testing.T) {
		fileContent, err := fs.Get(ctx, cfg.S3Endpoint, testFile.Name())
		if err != nil {
			t.Error(err)
		}
		if !bytes.Equal(fileContent, testFileContent) {
			t.Error(storage.ErrFileEmpty)
		}
	})

	t.Run("Delete file from bucket", func(t *testing.T) {
		err := fs.Delete(ctx, cfg.S3Endpoint, testFile.Name())
		if err != nil {
			t.Error(err)
		}
		_, err = fs.Get(ctx, cfg.S3Endpoint, testFile.Name())
		if err == nil {
			t.Error(err)
		}
	})
}
