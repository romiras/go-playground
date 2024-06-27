package main

import (
	"fmt"

	"github.com/mind1949/googletrans"
	"golang.org/x/text/language"
)

func main() {
	serviceURLs := []string{
		"https://translate.google.com/",
		"https://translate.google.com.br/"}
	googletrans.Append(serviceURLs...)

	params := googletrans.TranslateParams{
		Src:  "en", // "auto"
		Dest: language.Spanish.String(),
		Text: "Go is an open source programming language that makes it easy to build simple, reliable, and efficient software. ",
	}
	translated, err := googletrans.Translate(params)
	if err != nil {
		panic(err)
	}
	fmt.Printf("text: %q \npronunciation: %q", translated.Text, translated.Pronunciation)
}
