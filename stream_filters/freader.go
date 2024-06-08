// credits: https://play.golang.org/p/rakSV5kgqR9 @uvelichitel

package main

import (
	"io"
)

type (
	FReader struct {
		reader io.Reader
		filter R
	}

	R func([]byte) (int, error)
)

func NewFReader(r io.Reader) *FReader {
	return &FReader{r, func(p []byte) (int, error) { return len(p), nil }}
}

func (fr *FReader) Filter(filter R) *FReader {
	return &FReader{fr, filter}
}
func (fr *FReader) Read(p []byte) (n int, err error) {
	n, err = fr.reader.Read(p)
	if err == nil && fr.filter != nil {
		n, err = fr.filter(p)
	}
	return
}
