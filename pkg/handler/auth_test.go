package handler_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/rohitxdev/go-api-template/pkg/config"
	"github.com/rohitxdev/go-api-template/pkg/handler"
	"github.com/rohitxdev/go-api-template/pkg/repo"
)

func TestAuth(t *testing.T) {
	c, err := config.Load("../.env")
	if err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("postgres", c.DB_URL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	r := repo.New(db)
	h := handler.New(c, r, nil)
	e, err := handler.NewRouter(h)
	if err != nil {
		t.Fatal(err)
	}

	var accessToken string
	var refreshTokenCookie *http.Cookie
	t.Run("Sign up", func(t *testing.T) {
		req, err := createHttpRequest(http.MethodPost, "/v1/auth/sign-up", nil, echo.Map{"email": "user@test.com", "password": "password123", "confirm_password": "password123"}, map[string]string{"Content-Type": "application/json"})
		if err != nil {
			t.Error(err)
		}
		res := httptest.NewRecorder()
		c := e.NewContext(req, res)
		_ = h.SignUp(c)
		if res.Code != http.StatusCreated {
			t.Error(res.Body.String())
		}
		var data handler.LogInResponse
		_ = json.Unmarshal(res.Body.Bytes(), &data)
		accessToken = data.AccessToken
		for _, cookie := range res.Result().Cookies() {
			refreshTokenCookie = cookie
		}
	})

	t.Run("Log in", func(t *testing.T) {
		if accessToken == "" || refreshTokenCookie == nil {
			t.Skip()
		}
		req, err := createHttpRequest(http.MethodPost, "/v1/auth/log-in", nil, echo.Map{"email": "user@test.com", "password": "password123"}, map[string]string{"Content-Type": "application/json"})
		if err != nil {
			t.Error(err)
		}
		res := httptest.NewRecorder()
		c := e.NewContext(req, res)
		_ = h.LogIn(c)
		if res.Code != http.StatusOK {
			t.Error(res.Body.String())
		}
		var data handler.LogInResponse
		_ = json.Unmarshal(res.Body.Bytes(), &data)
		accessToken = data.AccessToken
		refreshTokenCookie = nil
		for _, cookie := range res.Result().Cookies() {
			refreshTokenCookie = cookie
		}
	})

	t.Run("Log out", func(t *testing.T) {
		if accessToken == "" || refreshTokenCookie == nil {
			t.Skip()
		}
		req, err := createHttpRequest(http.MethodPost, "/v1/auth/log-out", nil, nil, map[string]string{"Authorization": fmt.Sprintf("Bearer %s", accessToken), "Content-Type": "application/json"})
		if err != nil {
			t.Error(err)
		}
		req.AddCookie(refreshTokenCookie)
		res := httptest.NewRecorder()
		c := e.NewContext(req, res)
		_ = h.LogOut(c)
		if res.Code != http.StatusOK {
			t.Error(res.Body.String())
		}
	})

	t.Run("Delete account", func(t *testing.T) {
		if accessToken == "" {
			t.Skip()
		}
		req, err := createHttpRequest(http.MethodDelete, "/v1/auth/delete-account", nil, nil, map[string]string{"Authorization": fmt.Sprintf("Bearer %s", accessToken), "Content-Type": "application/json"})
		if err != nil {
			t.Error(err)
		}
		res := httptest.NewRecorder()
		authMiddleware := h.Auth(handler.RoleUser)
		c := e.NewContext(req, res)
		_ = authMiddleware(h.DeleteAccount)(c)
		if res.Code != http.StatusOK {
			t.Error(res.Body.String())
		}
	})
}

func createHttpRequest(method, path string, query map[string]string, body echo.Map, headers map[string]string) (*http.Request, error) {
	url, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	q := url.Query()
	for key, value := range query {
		q.Set(key, value)
	}
	url.RawQuery = q.Encode()
	j, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req := httptest.NewRequest(method, url.String(), bytes.NewReader(j))
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return req, err
}
