package services_test

import (
	"testing"

	"github.com/rohitxdev/go-api-template/internal/services"
)

func TestAccessToken(t *testing.T) {
	if err := services.GetAccessToken(); err == "" {
		t.Error("Failed")
	}

}
