package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	gin_context "github.com/repo/gin-file-uploader/internal/context"
	"github.com/repo/gin-file-uploader/internal/utils"
)

func UploadFileHandler(c *gin.Context) {
	code, err := uploadFile(gin_context.NewGinContext(c))
	if err != nil {
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func uploadFile(ginCtx *gin_context.GinContext) (code int, err error) {
	file, err := ginCtx.GetFormFile()
	if err != nil {
		return http.StatusBadRequest, err
	}

	dst := utils.GetLocalFilePath(file.Filename)
	log.Printf("Uploading file to %s", dst)

	err = ginCtx.SaveUploadedFile(file, dst)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}
