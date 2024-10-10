package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type getFileRequest struct {
	FileName string `param:"file_name" validate:"required"`
}

func (h *handler) GetFile(c echo.Context) error {
	req := new(getFileRequest)
	if err := bindAndValidate(c, req); err != nil {
		return err
	}
	presignedReq, err := h.blobstore.PresignGetObject(c.Request().Context(), h.config.S3BucketName, req.FileName)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, presignedReq)
}

type putFileRequest struct {
	File string `form:"file" validate:"required"`
}

func (h *handler) PutFile(c echo.Context) error {
	req := new(putFileRequest)
	if err := bindAndValidate(c, req); err != nil {
		return err
	}
	file, err := c.FormFile("file")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	f, err := file.Open()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer f.Close()
	// fileContent, err := io.ReadAll(f)
	// err = h.fs.Upload(c.Request().Context(), h.config.S3BucketName, file.Filename, fileContent)
	// if err != nil {
	// 	return c.String(http.StatusInternalServerError, err.Error())
	// }
	return c.NoContent(http.StatusOK)
}

func (h *handler) GetFileList(c echo.Context) error {
	files, err := h.blobstore.GetList(c.Request().Context(), h.config.S3BucketName, "")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, files)
}
