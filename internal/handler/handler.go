package handler

import (
	"embed"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rohitxdev/go-api-starter/internal/config"
	"github.com/rohitxdev/go-api-starter/pkg/blobstore"
	"github.com/rohitxdev/go-api-starter/pkg/email"
	"github.com/rohitxdev/go-api-starter/pkg/repo"
	"github.com/rohitxdev/go-api-starter/pkg/sqlite"
)

type Opts struct {
	Config   *config.Server
	Kv       *sqlite.KV
	Repo     *repo.Repo
	Email    *email.Client
	Fs       *blobstore.Store
	StaticFS *embed.FS
}

// type Opts struct {
// 	Config      *config.Server   // Configuration server
// 	KeyValue    *sqlite.KV       // Key-value store abstraction
// 	Repository  *repo.Repo       // Data repository
// 	EmailClient *email.Client    // Email service client
// 	FileStore   *blobstore.Store // Blob store for file storage
// 	StaticFiles *embed.FS        // Embedded static file system
// }

type handler struct {
	config   *config.Server
	kv       *sqlite.KV
	repo     *repo.Repo
	email    *email.Client
	fs       *blobstore.Store
	staticFS *embed.FS
}

// bindAndValidate binds path params, query params and the request body into provided type `i` and validates provided `i`. The default binder binds body based on Content-Type header. Validator must be registered using `Echo#Validator`.
func bindAndValidate(c echo.Context, i any) error {
	var err error
	if err = c.Bind(i); err != nil {
		_ = c.String(http.StatusInternalServerError, err.Error())
		return err
	}
	binder := echo.DefaultBinder{}
	if err = binder.BindHeaders(c, i); err != nil {
		_ = c.String(http.StatusInternalServerError, err.Error())
		return err
	}
	if err = c.Validate(i); err != nil {
		_ = c.String(http.StatusUnprocessableEntity, err.Error())
		return err
	}
	return err
}

func SanitizeEmail(email string) string {
	emailParts := strings.Split(email, "@")
	username := emailParts[0]
	domain := emailParts[1]
	if strings.Contains(username, "+") {
		username = strings.Split(username, "+")[0]
	}
	username = strings.ReplaceAll(username, "-", "")
	username = strings.ReplaceAll(username, ".", "")
	return username + "@" + domain
}
