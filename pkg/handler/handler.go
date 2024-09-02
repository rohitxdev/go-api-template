package handler

import (
	"embed"

	"github.com/rohitxdev/go-api-template/pkg/config"
	"github.com/rohitxdev/go-api-template/pkg/kv"
	"github.com/rohitxdev/go-api-template/pkg/repo"
	"github.com/rohitxdev/go-api-template/pkg/service"
)

type Handler struct {
	config   *config.Server
	kv       *kv.KV
	repo     *repo.Repo
	email    *service.EmailClient
	fs       *service.FileStorage
	staticFS *embed.FS
}

func New(c *config.Server, kv *kv.KV, r *repo.Repo, staticFS *embed.FS) *Handler {
	return &Handler{
		config:   c,
		kv:       kv,
		repo:     r,
		staticFS: staticFS,
	}
}
