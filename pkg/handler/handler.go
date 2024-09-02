package handler

import (
	"embed"

	"github.com/rohitxdev/go-api-template/pkg/config"
	"github.com/rohitxdev/go-api-template/pkg/repo"
	"github.com/rohitxdev/go-api-template/pkg/service"
	"github.com/rohitxdev/go-api-template/pkg/sqlite"
)

type Handler struct {
	config   *config.Server
	kv       *sqlite.KV
	repo     *repo.Repo
	email    *service.EmailClient
	fs       *service.FileStorage
	staticFS *embed.FS
}

func New(c *config.Server, kv *sqlite.KV, r *repo.Repo, staticFS *embed.FS) *Handler {
	return &Handler{
		config:   c,
		kv:       kv,
		repo:     r,
		staticFS: staticFS,
	}
}
