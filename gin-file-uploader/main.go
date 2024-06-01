package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/repo/gin-file-uploader/internal/handlers"
	"github.com/repo/gin-file-uploader/internal/utils"
)

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.PUT("/upload", handlers.UploadFile)

	return r
}

func init() {
	err := utils.CreateUploadsDir(utils.UploadsDirName)
	if err != nil {
		log.Fatal("Failed to create directory for uploads:", err)
	}
}

func main() {
	r := setupRouter()
	r.Run(":8080")
}
