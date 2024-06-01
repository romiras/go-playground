package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/repo/gin-file-uploader/internal/utils"
)

func UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dst := utils.GetLocalFilePath(file.Filename)
	log.Printf("Uploading to %s", dst)

	err = c.SaveUploadedFile(file, dst)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully", "file": dst})
}
