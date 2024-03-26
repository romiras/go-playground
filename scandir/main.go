package main

import (
	"encoding/csv"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

type FileState struct {
	Path string
	Size int64
}

func scanDir(path string, callback func(FileState)) error {
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			callback(FileState{
				Path: path,
				Size: info.Size(),
			})
		}
		return nil
	})
}

func main() {
	if len(os.Args) == 0 {
		log.Fatal("Example: scandir <dir> <output.log>")
	}

	file, err := os.OpenFile(os.Args[2], os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = scanDir(os.Args[1], func(state FileState) {
		err := writer.Write([]string{state.Path, strconv.FormatInt(state.Size, 10)})
		if err != nil {
			log.Fatal(err)
		}
	})
	if err != nil {
		log.Fatal(err)
	}
}
