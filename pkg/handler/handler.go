package handler

import (
	"embed"

	"github.com/rohitxdev/go-api-template/pkg/config"
	"github.com/rohitxdev/go-api-template/pkg/repo"
	"github.com/rohitxdev/go-api-template/pkg/service"
)

type Handler struct {
	config   *config.Server
	repo     *repo.Repo
	email    *service.EmailClient
	fs       *service.FileStorage
	staticFS *embed.FS
}

func New(c *config.Server, r *repo.Repo, staticFS *embed.FS) *Handler {
	return &Handler{
		config:   c,
		repo:     r,
		staticFS: staticFS,
	}
}
