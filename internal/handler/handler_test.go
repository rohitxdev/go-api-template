package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/rohitxdev/go-api-template/internal/handler"
)

func TestLogOut(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/log-out", nil)
	rec := httptest.NewRecorder()

	// assert.Equal(t, "Logged out", rec.Body.String(), "Should be 400")
	t.Run("Log out", func(t *testing.T) {
		c := echo.New().NewContext(req, rec)
		handler.LogOut(c)
		if rec.Body.String() == "Logged out" {
			t.Fail()
		}
	})
}
