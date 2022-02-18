package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

type BlobReader struct {
	keys       []string
	currentKey uint64
	reader     io.ReadCloser
}

func NewBlobReader(keys []string) *BlobReader {
	return &BlobReader{
		keys: keys,
	}
}

// Read реализация интерфейса io.Reader, чтобы можно было reader использовать в io.Copy
func (br *BlobReader) Read(p []byte) (int, error) {
	var err error

	// открываем каждый файл по очереди
	if br.reader == nil {
		filePath := filepath.Join(".", "blobs", br.keys[br.currentKey])
		br.reader, err = os.Open(filePath)
		if err != nil {
			br.reader.Close()
			return 0, err
		}
	}

	// читаем данные из текущего открытого файла, пока данные не закончатся
	n, err := br.reader.Read(p)

	// если данные в файле закончились, закрываем его
	if err == io.EOF {
		br.currentKey++
		br.reader.Close()
		br.reader = nil

		// io.EOF в err должно вернуть только у последнего файла, чтобы io.Copy считал все файлы и не завис на последнем.
		if br.currentKey < uint64(len(br.keys)) {
			err = nil
		}
	}

	return n, err
}

func (br *BlobReader) Close() error {
	if br.reader == nil {
		return nil
	}
	return br.reader.Close()
}

func main() {
	keys := []string{
		"2050-part1",
		"2050-part2",
		"2050-part3",
	}

	blobReader := NewBlobReader(keys)
	defer func() {
		_ = blobReader.Close()
	}()

	_, err := io.Copy(os.Stdout, blobReader)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Done")
}
