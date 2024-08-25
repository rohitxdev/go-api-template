package service_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/rohitxdev/go-api-template/pkg/config"
	"github.com/rohitxdev/go-api-template/pkg/service"
)

func TestStorageService(t *testing.T) {
	c, err := config.Load("../.env")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	fs, err := service.NewFileStorage(c.S3_ENDPOINT, c.S3_DEFAULT_REGION, c.AWS_ACCESS_KEY_ID, c.AWS_ACCESS_KEY_SECRET)
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
		err := fs.Upload(ctx, c.S3_ENDPOINT, testFile.Name(), testFileContent)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("Get file from bucket", func(t *testing.T) {
		fileContent, err := fs.Get(ctx, c.S3_ENDPOINT, testFile.Name())
		if err != nil {
			t.Error(err)
		}
		if !bytes.Equal(fileContent, testFileContent) {
			t.Error(service.ErrFileEmpty)
		}
	})

	t.Run("Delete file from bucket", func(t *testing.T) {
		err := fs.Delete(ctx, c.S3_ENDPOINT, testFile.Name())
		if err != nil {
			t.Error(err)
		}
		_, err = fs.Get(ctx, c.S3_ENDPOINT, testFile.Name())
		if err == nil {
			t.Error(err)
		}
	})
}
