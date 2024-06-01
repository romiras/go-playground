package utils

import (
	"os"
	"path/filepath"
)

const UploadsDirName = "_uploads"

func CreateUploadsDir(dirName string) error {
	dir := filepath.Dir(dirName)

	// Create the directory if it doesn't exist
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetLocalFilePath(filename string) string {
	return filepath.Join(UploadsDirName, filepath.Base(filename))
}
