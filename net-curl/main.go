package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func curl(url string) string {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", bodyText)

	return ""
}

func main() {
	fmt.Println(curl("https://en.wikipedia.org/robots.txt"))
}
