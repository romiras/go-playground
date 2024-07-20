package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"github.com/tmc/langchaingo/textsplitter"
)

const (
	RecursiveCharacterMode = iota
	TokenSplitterMode
	MarkdownTextSplitterMode
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: text_chunks <splitter_mode>")
		fmt.Println("Splitter modes: 0 (RecursiveCharacter), 1 (TokenSplitter), 2 (MarkdownTextSplitter)")
		return
	}

	mode, err := strconv.Atoi(os.Args[1])
	if err != nil || mode < 0 || mode > 2 {
		fmt.Println("Invalid splitter mode. Please use 0, 1, or 2.")
		return
	}

	text, err := readFromStdin()
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		return
	}

	splitter := selectSplitter(mode)

	if err := printTextChunks(splitter, text); err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
}

func readFromStdin() (string, error) {
	var text string

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text += scanner.Text() + "\n"
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading from stdin: %s\n", err.Error())
		return "", err
	}

	return text, nil
}

func selectSplitter(mode int) textsplitter.TextSplitter {
	switch mode {
	case RecursiveCharacterMode:
		return textsplitter.NewRecursiveCharacter()
	case TokenSplitterMode:
		return textsplitter.NewTokenSplitter(textsplitter.WithChunkSize(512), textsplitter.WithChunkOverlap(50))
	case MarkdownTextSplitterMode:
		return textsplitter.NewMarkdownTextSplitter()
	}

	return nil
}

func printTextChunks(splitter textsplitter.TextSplitter, text string) error {
	chunks, err := splitter.SplitText(text)
	if err != nil {
		return err
	}

	for _, chunk := range chunks {
		fmt.Println("---------------")
		fmt.Println(chunk)
	}

	return nil
}
