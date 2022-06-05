package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

func Usage() {
	log.Fatal("Usage: main.go <file.xml>")
}

func main() {
	if len(os.Args) != 2 {
		Usage()
	}

	fName := os.Args[1]
	if fName == "" || ".xml" != filepath.Ext(fName) {
		Usage()
	}

	file, err := os.Open(fName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	decoder := xml.NewDecoder(file)
	for {
		t, err := decoder.Token()      //  wrapper of the File Reader
		if t == nil || err == io.EOF { // when t is nil, we finished reading the file
			break // break the for
		}
		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "field" {
				var row string
				err = decoder.DecodeElement(&row, &se)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("%v: %+v\n", se.Attr[0].Value, row)
			}
		case xml.EndElement:
			if se.Name.Local == "row" {
				fmt.Println("")
			}
		default:
		}
	}
}
