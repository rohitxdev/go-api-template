package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/rohitxdev/go-api-template/internal/handlers"
	"github.com/stretchr/testify/assert"
)

func TestLogOut(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/log-out", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	handlers.LogOut(c)
	assert.Equal(t, "Logged out", rec.Body.String(), "Should be 400")
}
