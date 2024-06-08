package context

import (
	"mime/multipart"
)

type UploaderContext interface {
	GetFormFile() (*multipart.FileHeader, error)
	SaveUploadedFile(file *multipart.FileHeader, dst string) error
}
