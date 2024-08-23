package handler

import (
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rohitxdev/go-api-template/config"
	"github.com/rohitxdev/go-api-template/repo"
	"github.com/rohitxdev/go-api-template/service"
)

type Handler struct {
	config *config.Config
	repo   *repo.Repo
	email  *service.EmailClient
	fs     *service.FileStorage
}

func NewHandler(c *config.Config, r *repo.Repo) *Handler {
	handler := new(Handler)
	handler.config = c
	handler.repo = r
	return handler
}

func (h *Handler) GetUser(c echo.Context) error {
	user, err := h.repo.GetUserById(c.Request().Context(), 0)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, *user)
}

func (h *Handler) GetQuote(c echo.Context) error {
	res, err := http.DefaultClient.Get("https://api.kanye.rest")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, body)
}
