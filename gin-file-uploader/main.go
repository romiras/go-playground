package main

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func uploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dst := filepath.Join("_uploads", filepath.Base(file.Filename))
	log.Printf("Uploading to %s", dst)

	err = c.SaveUploadedFile(file, dst)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully", "file": dst})
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.PUT("/upload", uploadFile)

	return r
}

func main() {
	r := setupRouter()
	r.Run(":8080")
}
