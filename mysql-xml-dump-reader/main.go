package main

import (
	"encoding/csv"
	"encoding/xml"
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

	w := csv.NewWriter(os.Stdout)
	defer w.Flush()
	records := make([]string, 0, 4)

	decoder := xml.NewDecoder(file)
	for {
		t, err := decoder.Token()      //  wrapper of the File Reader
		if t == nil || err == io.EOF { // when t is nil, we finished reading the file
			break // break the for
		}
		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "field" {
				var value string
				err = decoder.DecodeElement(&value, &se)
				if err != nil {
					log.Fatal(err)
				}
				records = append(records, value)
			}
		case xml.EndElement:
			if se.Name.Local == "row" {
				w.Write(records)
				records = records[:0] // reset the slice
				w.Flush()
			}
		default:
		}
	}
}
