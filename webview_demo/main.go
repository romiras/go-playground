package main

import (
	"flag"
	"github.com/webview/webview"
)

const DefaultURL string = "https://en.m.wikipedia.org/wiki/Main_Page"

func main() {
	url := flag.String("url", DefaultURL, "a URL to to navigate")
	title := flag.String("title", "Minimal webview example", "a title of window")
	flag.Parse()

	debug := false
	w := webview.New(debug)
	defer w.Destroy()
	w.SetTitle(*title)
	w.SetSize(800, 600, webview.HintNone)
	w.Navigate(*url)
	w.Run()
}
