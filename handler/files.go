package handler

import (
	"io"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/rohitxdev/go-api-template/config"
	"github.com/rohitxdev/go-api-template/service"
	"github.com/rohitxdev/go-api-template/util"
)

type getFileRequest struct {
	FileName string `param:"file_name" validate:"required"`
}

func GetFile(c echo.Context) error {
	req := new(getFileRequest)
	if err := util.BindAndValidate(c, req); err != nil {
		return err
	}
	file, err := service.GetFileFromBucket(c.Request().Context(), config.S3_BUCKET_NAME, req.FileName)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.Blob(http.StatusFound, http.DetectContentType(file), file)
}

type putFileRequest struct {
	File string `form:"file" validate:"required"`
}

func PutFile(c echo.Context) error {
	req := new(putFileRequest)
	if err := util.BindAndValidate(c, req); err != nil {
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
	fileContent, err := io.ReadAll(f)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	err = service.UploadFileToBucket(c.Request().Context(), config.S3_BUCKET_NAME, file.Filename, fileContent)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}

func GetFileList(c echo.Context) error {
	files, err := service.GetFileList(c.Request().Context(), config.S3_BUCKET_NAME, "")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, files)
}
