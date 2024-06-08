package context

import (
	"mime/multipart"

	"github.com/gin-gonic/gin"
)

type GinContext struct {
	ctx *gin.Context
}

func NewGinContext(ctx *gin.Context) UploaderContext {
	return &GinContext{
		ctx: ctx,
	}
}

func (gctx *GinContext) GetFormFile() (*multipart.FileHeader, error) {
	return gctx.ctx.FormFile("file")
}

func (gctx *GinContext) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	return gctx.ctx.SaveUploadedFile(file, dst)
}
