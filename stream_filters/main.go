package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

func capitalize(p []byte) (int, error) {
	for i := range p {
		p[i] = byte(unicode.ToUpper(rune(p[i])))
	}
	return len(p), nil
}
func spare(p []byte) (int, error) {
	var b []byte
	for _, v := range p {
		b = append(b, v, 32)
	}
	n := copy(p, b)
	return n, nil
}
func output(p []byte) (int, error) {
	fmt.Println(string(p))
	return len(p), nil
}

func applyFilters(r io.Reader) (io.Reader, error) {
	rOut, w := io.Pipe()

	filter := func(rd io.Reader, wr io.Writer) error {
		return nil
	}

	err := filter(r, w)
	if err != nil {
		return nil, err
	}

	return rOut, nil
}

var ansiText string = `color.txt
[0m[01;34mfilters[0m
[01;31mfilters-ansi.tgz[0m
go.mod
go.sum
main.go
Readme.md
registry.go
streamhandler.go
`

func main() {
	reader := strings.NewReader(ansiText)
	// reader := strings.NewReader("Hello, playground")
	// r := NewFReader(reader).Filter(output).Filter(capitalize).Filter(output).Filter(spare)

	r, err := applyFilters(reader)
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(os.Stdout, r)
	if err != nil {
		panic(err)
	}
}
