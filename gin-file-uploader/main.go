package main

import (
	"github.com/gin-gonic/gin"
	"github.com/repo/gin-file-uploader/internal/handlers"
)

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.PUT("/upload", handlers.UploadFile)

	return r
}

func main() {
	r := setupRouter()
	r.Run(":8080")
}
