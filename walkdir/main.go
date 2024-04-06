package main

import (
	"log"
	"os"
	"path/filepath"
)

func walkDir(path string, callback func(string)) error {
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			callback(path)
		}
		return nil
	})
}

func main() {
	if len(os.Args) == 0 {
		log.Fatal("Example: walkdir <dir> <output.log>")
	}

	file, err := os.OpenFile(os.Args[2], os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	err = walkDir(os.Args[1], func(path string) {
		_, err := file.WriteString(path + "\n")
		if err != nil {
			log.Fatal(err)
		}
	})
	if err != nil {
		log.Fatal(err)
	}
}
