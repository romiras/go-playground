package main

import (
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()

	file, err := os.Open("app.log")
	if err != nil {
		logger.Fatal(err.Error())
	}
	defer file.Close()

	bytesChan := make(chan []byte)
	go func() {
		readFile(file, bytesChan)
		close(bytesChan)
	}()

	for bytes := range bytesChan {
		fmt.Print(string(bytes))
	}
	fmt.Println()
}

func readFile(reader io.Reader, bytesChan chan []byte) {
	bytes := make([]byte, 1024)
	for {
		_, err := reader.Read(bytes)
		if err != nil {
			break
		}
		bytesChan <- bytes
	}
}
