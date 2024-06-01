package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/repo/gin-file-uploader/internal/utils"
)

func UploadFileHandler(c *gin.Context) {
	code, err := uploadFile(c)
	if err != nil {
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func uploadFile(c *gin.Context) (code int, err error) {
	file, err := c.FormFile("file")
	if err != nil {
		return http.StatusBadRequest, err
	}

	dst := utils.GetLocalFilePath(file.Filename)
	log.Printf("Uploading file to %s", dst)

	err = c.SaveUploadedFile(file, dst)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}
